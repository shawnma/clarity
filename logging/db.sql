CREATE database clarity;

CREATE TABLE if not exists LOG (
    Id bigint(20) NOT NULL AUTO_INCREMENT,
    RemoteAddr varchar(32),
    Method varchar(10),
    RequestContentType varchar(50),
    RequestLength int,
    RequestBody Text,
    ResponseCode int,
    ResponseContentType varchar(50),
    ResponseLength int,
    ResponseBody Text,
    Title varchar(400),
    URL varchar(1024),
    LogTime datetime NOT NULL,
    primary key(Id, LogTime)
) ENGINE = InnoDB DEFAULT CHARSET = utf8

CREATE USER shawn identified by 'xxx';
grant all privileges on clarity.* to shawn;