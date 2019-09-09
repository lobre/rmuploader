# rmuploader
Web app to upload files to the Remarkable Tablet

## How to launch it

Get a one time code at https://my.remarkable.com/connect/remarkable.
Then launch it with Docker.

    docker run -ti --rm --name rmuploader -p 8080:8080 -e RMUPLOADER_CODE=<onetimecode> lobre/rmuploader

## Screenshot

![screenshot](https://raw.githubusercontent.com/lobre/rmuploader/master/screenshot.png)
