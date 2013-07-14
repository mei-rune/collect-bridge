
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS redis_commands;
DROP TABLE IF EXISTS actions;
DROP TABLE IF EXISTS metric_triggers;
DROP TABLE IF EXISTS triggers;
DROP TABLE IF EXISTS wbem_params;
DROP TABLE IF EXISTS ssh_params;
DROP TABLE IF EXISTS snmp_params;
DROP TABLE IF EXISTS endpoint_params;
DROP TABLE IF EXISTS access_params;
DROP TABLE IF EXISTS addresses;
DROP TABLE IF EXISTS interfaces;
DROP TABLE IF EXISTS links;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS managed_objects;
DROP TABLE IF EXISTS attributes;

CREATE TABLE managed_objects (
  id          SERIAL PRIMARY KEY,
  name        varchar(250),
  description varchar(2000),
  created_at  timestamp,
  updated_at  timestamp
);

CREATE TABLE attributes (
  id                 SERIAL PRIMARY KEY,
  managed_object_id  integer,
  description        varchar(2000)
);

CREATE TABLE devices (
	id            SERIAL PRIMARY KEY,
  name          varchar(250),
  description   varchar(2000),
  created_at    timestamp,
  updated_at    timestamp,

  address       varchar(250),
  manufacturer  integer,
  catalog       integer,
  oid           varchar(250),
  services      integer,
  location      varchar(2000),
  type          varchar(100)
);


CREATE TABLE links (
	id                      SERIAL PRIMARY KEY,
  name                    varchar(250),
  description             varchar(2000),
  created_at              timestamp,
  updated_at              timestamp,

  custom_speed_up         integer,
  custom_speed_down       integer,
  device1                 integer,
  ifIndex1                integer,
  device2                 integer,
  ifIndex2                integer,
  sampling_direct         integer
);

CREATE TABLE  interfaces (
	id                      SERIAL PRIMARY KEY,
  name                    varchar(250),
  description             varchar(2000),
  created_at              timestamp,
  updated_at              timestamp,
  ifIndex                 integer,
  ifDescr                 varchar(2000),
  ifType                  integer,
  ifMtu                   integer,
  ifSpeed                 integer,
  ifPhysAddress           varchar(50),
  device_id               integer
) ;

CREATE TABLE addresses (
	id                      SERIAL PRIMARY KEY,
  name                    varchar(250),
  description             varchar(2000),
  created_at              timestamp,
  updated_at              timestamp,
  address                 varchar(50),
  ifIndex                 integer,
  netmask                 varchar(50),
  bcastAddress            integer,
  reasmMaxSize            integer,
  device_id               integer
);


CREATE TABLE snmp_params (
  id                 SERIAL PRIMARY KEY,
  managed_object_id  integer,

  port               integer,
  version            varchar(50),

  read_community VARCHAR(50),
  write_community VARCHAR(50),

  sec_model VARCHAR(50),    -- usm
  read_sec_name VARCHAR(50),
  read_auth_pass VARCHAR(50),
  read_priv_pass VARCHAR(50),

  write_sec_name VARCHAR(50),
  write_auth_pass VARCHAR(50),
  write_priv_pass VARCHAR(50),

  max_msg_size INTEGER,
  context_name VARCHAR(50),
  identifier VARCHAR(50),
  engine_id VARCHAR(50)

) ;

CREATE TABLE ssh_params (
  id                 SERIAL PRIMARY KEY,
  managed_object_id  integer,
  description        varchar(2000),

  address            varchar(50),
  port               integer,
  user_name          varchar(50),
  user_password      varchar(250)
);

CREATE TABLE wbem_params (
  id                 SERIAL PRIMARY KEY,
  managed_object_id  integer,
  description        varchar(2000),
  url                varchar(2000),
  user_name          varchar(50),
  user_password      varchar(250)
) ;


CREATE TABLE triggers (
  id            SERIAL PRIMARY KEY,
  name          varchar(250),
  expression    varchar(250),
  attachment    varchar(2000),
  description   varchar(2000),

  parent_type   varchar(250),
  parent_id     integer
);

CREATE TABLE metric_triggers (
  id                 SERIAL PRIMARY KEY,
  name               varchar(250),
  expression         varchar(250),
  attachment         varchar(2000),
  description        varchar(2000),

  parent_type        varchar(250),
  parent_id          integer,
  metric             varchar(250),
  managed_object_id  integer
) ;

CREATE TABLE actions (
  id                 SERIAL PRIMARY KEY,
  name               varchar(250),
  description        varchar(2000),
  parent_type        varchar(250),
  parent_id          integer
);

CREATE TABLE redis_commands (
  id                 SERIAL PRIMARY KEY,
  name               varchar(250),
  description        varchar(2000),
  parent_type        varchar(250),
  parent_id          integer,

  command            varchar(10),
  arg0               varchar(200),
  arg1               varchar(200),
  arg2               varchar(200),
  arg3               varchar(200),
  arg4               varchar(200)
);


CREATE TABLE alerts (
  id                 SERIAL PRIMARY KEY,
  name               varchar(250),
  description        varchar(2000),
  parent_type        varchar(250),
  parent_id          integer,

  delay_times       integer,
  expression_style   varchar(50),
  expression_code    varchar(2000)
);


DROP TABLE IF EXISTS documents;
create table documents (
  id                 SERIAL PRIMARY KEY,
  name               varchar(100),
  type               varchar(100), 
  page_count         integer, 
  author             varchar(100), 
  bytes              integer,
  journal_id         integer,
  isbn               varchar(100),
  compressed_format  varchar(10), 
  website_id         integer, 
  user_id            integer,
  printer_id         integer,
  publish_at         integer
);

DROP TABLE IF EXISTS websites;
CREATE TABLE websites (id  SERIAL PRIMARY KEY, url varchar(200));

DROP TABLE IF EXISTS printers;
CREATE TABLE printers (id  SERIAL PRIMARY KEY, name varchar(200));

DROP TABLE IF EXISTS topics;
CREATE TABLE topics (id  SERIAL PRIMARY KEY, name varchar(200));

-- tables for CLOB
DROP TABLE IF EXISTS zip_files;
CREATE TABLE zip_files (id  SERIAL PRIMARY KEY, body text, document_id integer);