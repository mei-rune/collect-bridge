

DROP TABLE IF EXISTS tpt_mails;
DROP TABLE IF EXISTS tpt_syslogs;
DROP TABLE IF EXISTS tpt_db_commands;
DROP TABLE IF EXISTS tpt_exec_commands;
DROP TABLE IF EXISTS tpt_notification_groups;
DROP TABLE IF EXISTS tpt_histories CASCADE;
DROP TABLE IF EXISTS tpt_alert_histories CASCADE;
DROP TABLE IF EXISTS tpt_alerts;
DROP TABLE IF EXISTS tpt_redis_commands;
DROP TABLE IF EXISTS tpt_actions CASCADE;
DROP TABLE IF EXISTS tpt_metric_triggers;
DROP TABLE IF EXISTS tpt_triggers CASCADE;
DROP TABLE IF EXISTS tpt_wbem_params;
DROP TABLE IF EXISTS tpt_ssh_params;
DROP TABLE IF EXISTS tpt_snmp_params;
DROP TABLE IF EXISTS tpt_endpoint_params;
DROP TABLE IF EXISTS tpt_access_params CASCADE;
DROP TABLE IF EXISTS tpt_network_links;
DROP TABLE IF EXISTS tpt_network_addresses;
DROP TABLE IF EXISTS tpt_network_device_ports;
DROP TABLE IF EXISTS tpt_network_devices CASCADE;
DROP TABLE IF EXISTS tpt_managed_objects CASCADE;
DROP TABLE IF EXISTS tpt_attributes CASCADE;
DROP SEQUENCE IF EXISTS tpt_actions_seq;
DROP SEQUENCE IF EXISTS tpt_triggers_seq;
DROP SEQUENCE IF EXISTS tpt_managed_object_seq;
DROP SEQUENCE IF EXISTS tpt_attributes_seq;


CREATE SEQUENCE tpt_managed_object_seq;

CREATE TABLE tpt_managed_objects (
  id integer NOT NULL DEFAULT nextval('tpt_managed_object_seq')  PRIMARY KEY,
  description varchar(2000),
  created_at timestamp,
  updated_at timestamp
);

CREATE SEQUENCE tpt_attributes_seq;

CREATE TABLE tpt_attributes (
  id integer NOT NULL DEFAULT nextval('tpt_attributes_seq')  PRIMARY KEY,
  managed_object_id integer
);

CREATE TABLE tpt_network_devices (
  -- id integer NOT NULL DEFAULT nextval('tpt_managed_object_seq')  PRIMARY KEY,

  type VARCHAR(200),
  name VARCHAR(200),
  zh_name VARCHAR(200),
  address VARCHAR(50) NOT NULL,
  memo VARCHAR(255),
  manufacturer VARCHAR(100),
  device_type INT NOT NULL,
  owner VARCHAR(50),
  location VARCHAR(200),
  oid VARCHAR(200),
  CONSTRAINT tpt_network_devices_pkey PRIMARY KEY (id),
  CONSTRAINT tpt_network_devices_name_uq unique(address)

) INHERITS (tpt_managed_objects);


CREATE TABLE tpt_network_device_ports (
  -- id integer NOT NULL DEFAULT nextval('tpt_managed_object_seq')  PRIMARY KEY,
  name VARCHAR(100),
  if_index INT NOT NULL,
  if_descr VARCHAR(255),
  if_type INT,
  if_mtu INT,
  if_speed INT,
  if_physAddress VARCHAR(255),
  device_id BIGINT NOT NULL,

  CONSTRAINT tpt_network_device_ports_pkey PRIMARY KEY (id),
  FOREIGN KEY(device_id) REFERENCES tpt_network_devices(id)

) INHERITS (tpt_managed_objects);


CREATE TABLE tpt_network_links (
  -- id integer NOT NULL DEFAULT nextval('tpt_managed_object_seq')  PRIMARY KEY,
  name VARCHAR(200) NOT NULL,
  memo VARCHAR(255),
  from_device BIGINT NOT NULL,
  from_if_index BIGINT,
  to_device BIGINT NOT NULL,
  to_if_index BIGINT,

  custom_speed_up         integer,
  custom_speed_down       integer,

  CONSTRAINT tpt_network_links_pkey PRIMARY KEY (id),
  FOREIGN KEY(from_device) REFERENCES tpt_network_devices(id),
  FOREIGN KEY(to_device) REFERENCES tpt_network_devices(id)

) INHERITS (tpt_managed_objects);



CREATE TABLE tpt_network_addresses (
  -- id integer NOT NULL DEFAULT nextval('managed_object_seq')  PRIMARY KEY,
  address varchar(50),
  ifIndex integer,
  netmask varchar(50),
  bcastAddress integer,
  reasmMaxSize integer,
  device_id integer,
  CONSTRAINT tpt_network_addresses_pkey PRIMARY KEY (id),
  FOREIGN KEY(device_id) REFERENCES tpt_network_devices(id)
) INHERITS (tpt_managed_objects);


CREATE TABLE tpt_access_params (
  -- id integer NOT NULL DEFAULT nextval('tpt_attributes_seq')  PRIMARY KEY,
  CONSTRAINT tpt_access_params_pkey PRIMARY KEY (id)
) INHERITS (tpt_attributes);


CREATE TABLE tpt_endpoint_params (
  -- id integer NOT NULL DEFAULT nextval('tpt_attributes_seq')  PRIMARY KEY,
  port integer,
  CONSTRAINT tpt_endpoint_params_pkey PRIMARY KEY (id)
) INHERITS (tpt_access_params);


CREATE TABLE tpt_snmp_params (
  -- id integer NOT NULL DEFAULT nextval('tpt_attributes_seq')  PRIMARY KEY,
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

  CONSTRAINT tpt_snmp_params_pk PRIMARY KEY (id),
  FOREIGN KEY(managed_object_id) REFERENCES tpt_network_devices(id)
) INHERITS (tpt_endpoint_params);

CREATE TABLE tpt_ssh_params (
  -- id integer NOT NULL DEFAULT nextval('tpt_attributes_seq')  PRIMARY KEY,
  user_name varchar(50),
  user_password varchar(250),
  CONSTRAINT tpt_ssh_params_pkey PRIMARY KEY (id),
  FOREIGN KEY(managed_object_id) REFERENCES tpt_network_devices(id)
) INHERITS (tpt_endpoint_params);


CREATE TABLE tpt_wbem_params (
  -- id integer NOT NULL DEFAULT nextval('tpt_attributes_seq')  PRIMARY KEY,
  url varchar(2000),
  user_name varchar(50),
  user_password varchar(250),

  CONSTRAINT tpt_wbem_params_pkey PRIMARY KEY (id),
  FOREIGN KEY(managed_object_id) REFERENCES tpt_network_devices(id)
) INHERITS (tpt_access_params);


CREATE SEQUENCE tpt_triggers_seq;
CREATE TABLE tpt_triggers (
  id            integer NOT NULL DEFAULT nextval('tpt_triggers_seq')  PRIMARY KEY,
  name          varchar(250),
  expression    varchar(250),
  attachment    varchar(2000),
  description   varchar(2000),
  created_at    timestamp,
  updated_at    timestamp
);

CREATE TABLE tpt_metric_triggers (
  -- id integer NOT NULL DEFAULT nextval('tpt_triggers_seq')  PRIMARY KEY,
  metric varchar(250),

  managed_object_id integer,
  CONSTRAINT tpt_metric_triggers_pkey PRIMARY KEY (id)
) INHERITS (tpt_triggers);

CREATE SEQUENCE tpt_actions_seq;
CREATE TABLE tpt_actions (
  id integer NOT NULL DEFAULT nextval('tpt_actions_seq')  PRIMARY KEY,
  name varchar(250),
  description   varchar(2000),

  parent_type varchar(250),
  parent_id integer, 
  
  created_at    timestamp,
  updated_at    timestamp
);

CREATE TABLE tpt_redis_commands (
  -- id integer NOT NULL DEFAULT nextval('actions_seq')  PRIMARY KEY,
  command varchar(10),
  arg0  varchar(200),
  arg1  varchar(200),
  arg2  varchar(200),
  arg3  varchar(200),
  arg4  varchar(200),
  
  CONSTRAINT tpt_redis_commands_pkey PRIMARY KEY (id)
) INHERITS (tpt_actions);

CREATE TABLE tpt_mails (
  from_address varchar(250),
  to_address varchar(250) NOT NULL,
  cc_address varchar(250),
  bcc_address varchar(250),
  subject varchar(250) NOT NULL,
  content_type varchar(50) NOT NULL,
  content varchar(2000) NOT NULL,

  CONSTRAINT tpt_mails_pkey PRIMARY KEY (id)
) INHERITS (tpt_actions);


CREATE TABLE tpt_syslogs (
  facility varchar(50) NOT NULL,
  severity varchar(50) NOT NULL,
  tag varchar(100),
  content varchar(2000) NOT NULL,

  CONSTRAINT tpt_syslogs_pkey PRIMARY KEY (id)
) INHERITS (tpt_actions);


CREATE TABLE tpt_db_commands (
  drv varchar(200) NOT NULL,
  url varchar(200) NOT NULL,
  script varchar(2000) NOT NULL,

  CONSTRAINT tpt_db_commands_pkey PRIMARY KEY (id)
) INHERITS (tpt_actions);

CREATE TABLE tpt_exec_commands (
  work_directory varchar(500) ,
  prompt varchar(250) ,
  command varchar(500) NOT NULL,

  CONSTRAINT tpt_exec_commands_pkey PRIMARY KEY (id)
) INHERITS (tpt_actions);

CREATE TABLE tpt_alerts (
  -- id integer NOT NULL DEFAULT nextval('actions_seq')  PRIMARY KEY,
  delay_times           integer,
  enabled               boolean,
  level                 integer,

  expression_style      varchar(50),
  expression_code       varchar(2000),
  notification_group_id integer,

  CONSTRAINT tpt_alerts_pkey PRIMARY KEY (id)
) INHERITS (tpt_actions);

CREATE TABLE tpt_histories (
  enabled    boolean,
  attribute  varchar(200),
  CONSTRAINT tpt_histories_pkey PRIMARY KEY (id)
) INHERITS (tpt_actions);


CREATE TABLE tpt_notification_groups (
  id serial PRIMARY KEY,
  name varchar(200) NOT NULL,
  description varchar(2000),
  created_at timestamp,
  updated_at timestamp
);


DROP TABLE IF EXISTS tpt_documents;
create table tpt_documents (
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

DROP TABLE IF EXISTS tpt_websites;
CREATE TABLE tpt_websites (id  serial PRIMARY KEY, url varchar(200));

DROP TABLE IF EXISTS tpt_printers;
CREATE TABLE tpt_printers (id  serial PRIMARY KEY, name varchar(200));

DROP TABLE IF EXISTS tpt_topics;
CREATE TABLE tpt_topics (id  serial PRIMARY KEY, name varchar(200));

-- tables for CLOB
DROP TABLE IF EXISTS tpt_zip_files;
CREATE TABLE tpt_zip_files (id  serial PRIMARY KEY, body text, document_id integer);