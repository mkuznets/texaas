package workspace

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"mkuznets.com/go/texaas/internal/cache"
	"mkuznets.com/go/texaas/internal/fs"
	"mkuznets.com/go/texaas/internal/opts"
	"mkuznets.com/go/texaas/internal/workspace/pb"
)

type Service struct {
	cacheDir string
	toolsDir string
	docker   *opts.Docker
	output   *opts.Output
	pb.UnimplementedWorkspaceServer
}

func NewService(cacheDir, toolsDir string, docker *opts.Docker, output *opts.Output) *Service {
	return &Service{
		cacheDir: cacheDir,
		toolsDir: toolsDir,
		docker:   docker,
		output:   output,
	}
}

func (s *Service) basePath(id string, elems ...string) string {
	ps := []string{os.TempDir(), "texaas", fmt.Sprintf("ws_%s", id)}
	ps = append(ps, elems...)
	return filepath.Join(ps...)
}

func (s *Service) Get(stream pb.Workspace_GetServer) error {
	wsReq, err := stream.Recv()
	if err != nil {
		return err
	}

	defer func() {
		err := stream.Send(&pb.WS{Closed: true})
		log.Debug().Err(err).Msg("workspace closed")
	}()

	// -------------------------------------------------------------------------
	// Trying to lock existing or new workspace

	var (
		id  string
		mut *fs.Mutex
	)

	rnd := rand.NewSource(0xDEADBEEF)

	for i := 0; id == "" && i < 10; i++ {
		nonce := rnd.Int63()
		id = getID(wsReq, nonce)

		if err := os.MkdirAll(s.basePath(id), os.FileMode(0755)); err != nil {
			return err
		}

		mut = &fs.Mutex{Path: s.basePath(id, "lock")}
		ok, err := mut.Lock()
		if err != nil {
			return err
		}
		if !ok {
			id = ""
		}
	}

	if id == "" {
		return fmt.Errorf("could not lock workspace")
	}

	basePath := s.basePath(id)

	defer func() {
		log.Debug().Str("id", id).Msg("unlocking workspace")
		if err := mut.Unlock(); err != nil {
			log.Err(err).Str("id", id).Msg("could not release lock")
		}
	}()

	// -------------------------------------------------------------------------
	// Init workspace

	for _, name := range []string{"upper", "work", "merged"} {
		path := filepath.Join(basePath, name)
		if err := os.MkdirAll(path, os.FileMode(0755)); err != nil {
			return err
		}
	}

	outputDir := filepath.Join(basePath, "upper", "texaas", wsReq.Makefile.BasePath)
	outputLink := filepath.Join(basePath, "output")

	_ = os.Remove(outputLink)
	if err := os.Symlink(outputDir, outputLink); err != nil {
		return err
	}

	// -------------------------------------------------------------------------
	// Copy repo files

	repoDir := filepath.Join(basePath, "lower", "texaas")

	if err := os.RemoveAll(repoDir); err != nil {
		return err
	}
	if err := os.MkdirAll(repoDir, os.FileMode(0755)); err != nil {
		return err
	}

	defer func() {
		log.Debug().Str("id", id).Msg("removing repo files")
		if err := os.RemoveAll(repoDir); err != nil {
			log.Err(err).Str("id", id).Msg("could not remove repo files")
		}
	}()

	for _, file := range wsReq.Makefile.Inputs {
		item, err := cache.NewItem(file.Key)
		if err != nil {
			return err
		}
		src := item.Path(s.cacheDir)
		dst := filepath.Join(repoDir, file.Path)

		if err := os.MkdirAll(filepath.Dir(dst), os.FileMode(0755)); err != nil {
			return err
		}
		if err := fs.CopyFile(src, dst); err != nil {
			return err
		}
	}

	// -------------------------------------------------------------------------
	// Chown

	if err := fs.Chown(basePath, s.docker.UID, s.docker.GID); err != nil {
		return err
	}

	// -------------------------------------------------------------------------
	// Mount

	merged := filepath.Join(basePath, "merged")
	data := fmt.Sprintf(
		"lowerdir=%s,upperdir=%s,workdir=%s",
		filepath.Join(basePath, "lower"),
		filepath.Join(basePath, "upper"),
		filepath.Join(basePath, "work"),
	)

	if err := syscall.Mount("overlay", merged, "overlay", 0, data); err != nil {
		return err
	}

	defer func() {
		log.Debug().Str("id", id).Msg("unmounting overlay")
		if err := syscall.Unmount(merged, 0); err != nil {
			log.Err(err).Msg("could not unmount")
		}
	}()

	// -------------------------------------------------------------------------
	// Send workspace to client and wait for disconnect

	ws := &pb.WS{
		Path:  basePath,
		Tools: s.toolsDir,
		Id:    &pb.WSID{Id: id},
	}

	if err := stream.Send(ws); err != nil {
		return err
	}

	// Wait for client to finish
	_, err = stream.Recv()
	if err == io.EOF || stream.Context().Err() == context.Canceled {
		// Situation normal
		err = nil
	}

	log.Info().Err(err).Msg("disconnected, cleaning up workspace")

	return nil
}

func getID(req *pb.WSReq, nonce int64) string {
	s := fmt.Sprintf(
		"%s::%s::%s::%d",
		req.Makefile.MainSource,
		req.Makefile.Compiler,
		req.Makefile.Latex,
		nonce,
	)
	return fmt.Sprintf("%x", sha1.Sum([]byte(s)))
}

var expectedOutputs = []struct{ key, filename string }{
	{"pdf", "output.pdf"},
	{"log", "output.log"},
	{"latexmk", "latexmk.log"},
}

func outputPrefix() string {
	u, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%d-%s", time.Now().Unix(), u.String())
}

func (s *Service) Output(_ context.Context, wsID *pb.WSID) (*pb.WSOutput, error) {

	prefix := outputPrefix()

	dstBase := filepath.Join(s.output.Dir, prefix)
	if err := os.Mkdir(dstBase, os.FileMode(0755)); err != nil {
		return nil, err
	}

	base := s.basePath(wsID.Id, "output")

	output := &pb.WSOutput{
		Url:     s.output.URL,
		Outputs: make([]*pb.Output, 0),
	}

	for _, out := range expectedOutputs {
		src := filepath.Join(base, out.filename)
		dst := filepath.Join(dstBase, out.filename)

		fi, err := os.Stat(src)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}

		if err := fs.CopyFile(src, dst); err != nil {
			return nil, err
		}
		if err := os.Chmod(dst, os.FileMode(0400)); err != nil {
			return nil, err
		}

		output.Outputs = append(output.Outputs, &pb.Output{
			Key:  out.key,
			Path: filepath.Join(prefix, out.filename),
			Size: uint64(fi.Size()),
		})
	}

	if err := fs.Chown(dstBase, s.output.UID, s.output.GID); err != nil {
		return nil, err
	}

	return output, nil
}
