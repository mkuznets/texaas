$pdflatex = 'pdflatex -halt-on-error -interaction=batchmode -file-line-error %O %S '
. '&& echo %D > TARGET';
