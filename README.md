# rmuploader

Web app to upload files to the Remarkable Tablet.

## Upload PDF or EPUB

It aims to provide a simple way to drag and drop PDF and EPUB files to your Remarkable Tablet
from a Web interface. It can be useful when you don't have access to the official Remarkable application
or if you are running a Linux operating system (which does not have an official app).

It is written in Go and it makes use of the latest `api.v2` and `archive` packages introduced in `rmapi`. See:
 - github.com/juruen/rmapi/api.v2
 - github.com/juruen/rmapi/archive

## Send a webpage from a URL

This apps also provides an input to submit an URL. The webpage is crawled, converted into PDF and sent to the device.
It uses wkhtmltopdf underneath so it should be installed for this feature to work.

## How to launch the app

Get a one time code at https://my.remarkable.com/connect/remarkable.
Then launch it with Docker.

    docker run -ti --rm --name rmuploader -p 8080:8080 -e RMUPLOADER_CODE=<onetimecode> lobre/rmuploader

## Screenshot

![screenshot](https://raw.githubusercontent.com/lobre/rmuploader/master/screenshot.png)
