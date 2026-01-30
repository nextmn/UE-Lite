# Copyright Louis Royer and the NextMN contributors. All rights reserved.
# Use of this source code is governed by a MIT-style license that can be
# found in the LICENSE file.
# SPDX-License-Identifier: MIT

FROM golang:1.25.7 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
# To make reproducible the `COPY --from=builder` layer reproducible, we set modification time to build timestamp
# and we will copy the directory at once to avoid /usr/local/bin being "created" instead of copied (resulting in wrong a newer modification time).
RUN CGO_ENABLED=0 go build -trimpath -o /usr/local/bin/ue-lite && touch --no-dereference --date="@$(ue-lite --build-timestamp)" /usr/local/bin /usr/local/bin/ue-lite

FROM alpine:3.23.3
COPY --from=builder /usr/local/bin /usr/local/bin
# Even when cache is not created and logs are not written by apk-tools itself, adding some packages always updates modification time of a lot of files.
# To make this layer reprodicible, we need to reset the modification time to build timestamp.
# Some files are read-only filesystem (mounted by Docker), so we make sure to exclude them.
RUN apk add --no-cache --logfile=no iptables iproute2 && find /bin /etc /usr /lib /sbin /var -newer /usr/local/bin/ue-lite -not -path /etc/hosts -not path /etc/resolv.conf -print0 | xargs -0r touch --no-dereference --date="@$(ue-lite --build-timestamp)"
ENTRYPOINT ["ue-lite"]
CMD ["--help"]
HEALTHCHECK --interval=1m --timeout=1s --retries=3 --start-period=5s --start-interval=100ms \
CMD ["ue-lite", "healthcheck"]
