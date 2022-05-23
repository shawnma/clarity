# Clarity MITM

## General SSL certificate and key

openssl req -x509 -newkey rsa:4096 -nodes -keyout key.pem -out cert.pem -sha256 -days 3650 \
  -subj "/C=US/ST=California/L=Cupertino/O=Clarity/CN=Clarity CA"