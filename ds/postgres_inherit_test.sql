
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
DROP SEQUENCE IF EXISTS actions_seq;
DROP SEQUENCE IF EXISTS triggers_seq;
DROP SEQUENCE IF EXISTS managed_object_seq;
DROP SEQUENCE IF EXISTS attributes_seq;


CREATE SEQUENCE managed_object_seq;

CREATE TABLE managed_objects (
  id integer NOT NULL DEFAULT nextval('managed_object_seq')  PRIMARY KEY,
  name varchar(250),
  description varchar(2000),
  created_at timestamp,
  updated_at timestamp
);

CREATE SEQUENCE attributes_seq;

CREATE TABLE attributes (
  id integer NOT NULL DEFAULT nextval('attributes_seq')  PRIMARY KEY,
  managed_object_id integer
);

CREATE TABLE devices (
  -- id integer NOT NULL DEFAULT nextval('managed_object_seq')  PRIMARY KEY,
  address       varchar(250),
  manufacturer  integer,
  catalog       integer,
  oid           varchar(250),
  services      integer,
  location      varchar(2000),
  type          varchar(100), 
  CONSTRAINT devices_pkey PRIMARY KEY (id)
) INHERITS (managed_objects);


CREATE TABLE links (
  -- id integer NOT NULL DEFAULT nextval('managed_object_seq')  PRIMARY KEY,

  name                    varchar(250),
  custom_speed_up         integer,
  custom_speed_down       integer,
  device1                 integer,
  ifIndex1                integer,
  device2                 integer,
  ifIndex2                integer,
  sampling_direct         integer,
  CONSTRAINT links_pkey PRIMARY KEY (id)
) INHERITS (managed_objects);

CREATE TABLE  interfaces (
  -- id integer NOT NULL DEFAULT nextval('managed_object_seq')  PRIMARY KEY,
  ifIndex integer,
  ifDescr varchar(2000),
  ifType integer,
  ifMtu integer,
  ifSpeed integer,
  ifPhysAddress varchar(50),
  device_id integer,
  CONSTRAINT interfaces_pkey PRIMARY KEY (id)
) INHERITS (managed_objects);

CREATE TABLE addresses (
  -- id integer NOT NULL DEFAULT nextval('managed_object_seq')  PRIMARY KEY,
  address varchar(50),
  ifIndex integer,
  netmask varchar(50),
  bcastAddress integer,
  reasmMaxSize integer,
  device_id integer,
  CONSTRAINT addresses_pkey PRIMARY KEY (id)
) INHERITS (managed_objects);


CREATE TABLE access_params (
  -- id integer NOT NULL DEFAULT nextval('attributes_seq')  PRIMARY KEY,
  CONSTRAINT access_params_pkey PRIMARY KEY (id)
) INHERITS (attributes);


CREATE TABLE endpoint_params (
  -- id integer NOT NULL DEFAULT nextval('attributes_seq')  PRIMARY KEY,
  port integer,
  CONSTRAINT endpoint_params_pkey PRIMARY KEY (id)
) INHERITS (access_params);


CREATE TABLE snmp_params (
  -- id integer NOT NULL DEFAULT nextval('attributes_seq')  PRIMARY KEY,
  version VARCHAR(50),

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
  engine_id VARCHAR(50),

  CONSTRAINT snmp_params_pkey PRIMARY KEY (id)
) INHERITS (endpoint_params);

CREATE TABLE ssh_params (
  -- id integer NOT NULL DEFAULT nextval('attributes_seq')  PRIMARY KEY,
  user_name varchar(50),
  user_password varchar(250),
  CONSTRAINT ssh_params_pkey PRIMARY KEY (id)
) INHERITS (endpoint_params);

CREATE TABLE wbem_params (
  -- id integer NOT NULL DEFAULT nextval('attributes_seq')  PRIMARY KEY,
  url varchar(2000),
  user_name varchar(50),
  user_password varchar(250),
  CONSTRAINT wbem_params_pkey PRIMARY KEY (id)
) INHERITS (access_params);


CREATE SEQUENCE triggers_seq;
CREATE TABLE triggers (
  id            integer NOT NULL DEFAULT nextval('triggers_seq')  PRIMARY KEY,
  name          varchar(250),
  expression    varchar(250),
  attachment    varchar(2000),
  description   varchar(2000),

  parent_type   varchar(250),
  parent_id     integer
);

CREATE TABLE metric_triggers (
  -- id integer NOT NULL DEFAULT nextval('triggers_seq')  PRIMARY KEY,
  metric varchar(250),
  managed_object_id integer,
  CONSTRAINT metric_triggers_pkey PRIMARY KEY (id)
) INHERITS (triggers);

CREATE SEQUENCE actions_seq;
CREATE TABLE actions (
  id integer NOT NULL DEFAULT nextval('actions_seq')  PRIMARY KEY,
  name varchar(250),
  description   varchar(2000),

  parent_type varchar(250),
  parent_id integer
);

CREATE TABLE redis_commands (
  -- id integer NOT NULL DEFAULT nextval('actions_seq')  PRIMARY KEY,
  command varchar(10),
  arg0  varchar(200),
  arg1  varchar(200),
  arg2  varchar(200),
  arg3  varchar(200),
  arg4  varchar(200),
  
  CONSTRAINT redis_commands_pkey PRIMARY KEY (id)
) INHERITS (actions);


CREATE TABLE alerts (
  -- id integer NOT NULL DEFAULT nextval('actions_seq')  PRIMARY KEY,
  max_repeated  integer,
  expression_style varchar(50),
  expression_code  varchar(2000),
  
  CONSTRAINT alerts_pkey PRIMARY KEY (id)
) INHERITS (actions);


DROP TABLE IF EXISTS documents;
create table documents (
  id serial PRIMARY KEY,
  name varchar(100),
  type varchar(100), 
  page_count integer, 
  author varchar(100), 
  bytes integer,
  journal_id integer,
  isbn varchar(100),
  compressed_format varchar(10), 
  website_id integer, 
  user_id integer,
  printer_id integer,
  publish_at integer
);

DROP TABLE IF EXISTS websites;
CREATE TABLE websites (id  serial PRIMARY KEY, url varchar(200));

DROP TABLE IF EXISTS printers;
CREATE TABLE printers (id  serial PRIMARY KEY, name varchar(200));

DROP TABLE IF EXISTS topics;
CREATE TABLE topics (id  serial PRIMARY KEY, name varchar(200));

-- tables for CLOB
DROP TABLE IF EXISTS zip_files;
CREATE TABLE zip_files (id  serial PRIMARY KEY, body text, document_id integer);