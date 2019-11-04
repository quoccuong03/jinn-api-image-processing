#!/bin/bash
set -eo pipefail
shopt -s nullglob

# if command starts with an option, prepend imaginary
if [ "${1:0:1}" = '-' ]; then
	set -- /go/bin/server "$@"
fi

# usage: file_env VAR [DEFAULT]
#    ie: file_env 'XYZ_DB_PASSWORD' 'example'
# (will allow for "$XYZ_DB_PASSWORD_FILE" to fill in the value of
#  "$XYZ_DB_PASSWORD" from a file, especially for Docker's secrets feature)
file_env() {
	local var="$1"
	local fileVar="${var}_FILE"
	local def="${2:-}"
	if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
		echo >&2 "error: both $var and $fileVar are set (but are exclusive)"
		exit 1
	fi
	local val="$def"
	if [ "${!var:-}" ]; then
		val="${!var}"
	elif [ "${!fileVar:-}" ]; then
		val="$(< "${!fileVar}")"
	fi
	export "$var"="$val"
	unset "$fileVar"
}
file_env 'AWS_ACCESS_KEY_ID'
if [ -z "$AWS_ACCESS_KEY_ID" ]; then
			echo >&2 'error: S3 key has not set yet '
			echo >&2 '  You need to specify S3 key to upload images to S3'
			exit 1
fi
file_env 'AWS_SECRET_ACCESS_KEY'
if [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
			echo >&2 'error: S3 secret has not set yet '
			echo >&2 '  You need to specify S3 secret to upload images to S3'
			exit 1
fi

exec "$@"