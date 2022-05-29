# Clarity MITM

## General SSL certificate and key

openssl req -x509 -newkey rsa:4096 -nodes -keyout key.pem -out cert.pem -sha256 -days 3650 \
  -subj "/C=US/ST=California/L=Cupertino/O=Clarity/CN=Clarity CA"

##
{
    rule: DENY
    begin:
    end:
}
{
    rule: permissible
    begin:
    end:
    curl -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/email
    
}
{
    rule: TEMP
    begin:
    end:
}
aops.com/chat/emma/...

"" -> {}