# Private Podcast (pp)
Private Podcast is a simple application to serve podcast episodes from a S3 bucket with Google login.

## Building and Running
If you have the latest go toolchain installed running `go build ./cmd` should be enough.
To run the application you'll need to set the AWS environmental variables in addition to the configuration provided and documented on the CLI (see [cmd/main.go](cmd/main.go) for the variables and their documentation). The AWS variables that are usually needed are `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, the region should be the region of the S3 bucket.

## Database
This application requires a Postgres database, by default it connects to a local database. To run a local database for testing you can use `docker run -e POSTGRES_PASSWORD=secret -e POSTGRES_USER=pp -p 5432:5432 -it postgres:12`.

## S3 Bucket
**NEVER** change the key (name/path) of a podcast in S3, otherwise its GUID will also change, meaning that some podcast applications might show that particular episode multiple times. If you need to rename an episode you can add a metadata title (metadata with key of `x-amx-meta-title` in the S3 Console) to it. Publishing date cannot be changed currently.

All podcasts in the S3 bucket should be placed in the root, have a `.mp3` prefix, and start with the publishing date in YYYY-MM-DD format.
For example, a file named `2020-01-27 Hello World!.mp3` will be parsed as podcast episode that was released on the 27th of January in 2020, with a title and description of `Hello World!`. All files which can't be parsed are skipped.

To add a description to a podcast, another file can be added with `.txt` suffix. It's name must otherwise be exactly equal, e.g. in the example above the file would be named `2020-01-27 Hello World!.mp3.txt`.

## License

```
pp - Private Podcast feed with Single Sign-On
Copyright (C) 2020  Maximilian Remming

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
```
