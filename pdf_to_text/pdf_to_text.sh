#!/bin/sh
# This script creates word cloud from PDF file.
# provide one argument that is name of the pdf file
if [ "$1" ]
then
	echo "Name of the PDF file is provided"
	PDF_NAME="$1"
else
	echo "Provide the name of the PDF document: "
	read PDF_NAME
fi
# extract the words from PDF to text file
pdftotext "$PDF_NAME" all.txt
# remove non printable chars
perl -lpe s/[^[:print:]]+//g all.txt >> clean.txt

cat "clean.txt" | \
sed 's/[^a-zA-Z]/ /g' | \
tr '[:upper:]' '[:lower:]' | \
tr ' ' '
' | \
sed '/^$/d' | \
sed '/^[a-z]$/d' | \
grep -v -w -f stopwords | \
sort > words
#rm all.txt
#rm clean.txt

