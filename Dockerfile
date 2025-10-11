FROM surnet/alpine-wkhtmltopdf:3.20.2-0.12.6-full as builder

FROM golang:1.24-alpine

RUN mkdir /app
WORKDIR /app

# Install dependencies
RUN apk update && \
    apk add --no-cache \
        git \
        openssh \
        tzdata \
        build-base \
        python3 \
        net-tools \
        libstdc++ \
        libx11 \
        libxrender \
        libxext \
        libressl \
        ca-certificates \
        fontconfig \
        freetype \
        ttf-dejavu \
        ttf-droid \
        ttf-freefont \
        ttf-liberation \
    && apk add --no-cache --virtual .build-deps msttcorefonts-installer \
    && update-ms-fonts \
    && fc-cache -f \
    && rm -rf /var/cache/apk/* /tmp/* \
    && apk del .build-deps

# Copy source code
COPY .env.example .env
COPY . .

# Install tools & dependencies
RUN go install github.com/buu700/gin@latest
RUN go mod tidy

# Build binary
RUN make build

# Timezone setup
ENV TZ=Asia/Jakarta
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Copy wkhtmltopdf binaries from builder image
COPY --from=builder /usr/bin/wkhtmltopdf /usr/bin/wkhtmltopdf
COPY --from=builder /usr/bin/wkhtmltoimage /usr/bin/wkhtmltoimage

# Run app
EXPOSE 8003
ENTRYPOINT ["/app/payment-service"]
