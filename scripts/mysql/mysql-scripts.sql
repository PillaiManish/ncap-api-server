# TODO: add columns: created_timestamp, updated_timestamp

CREATE TABLE users (
  id int auto_increment,
  name varchar(250),
  email varchar(250) UNIQUE ,
  password varchar(250),
PRIMARY KEY (id)
);

CREATE TABLE project_details (
    id varchar(250),
    name varchar(250),
    description varchar(250)
);

CREATE TABLE project_metadata (
    id int AUTO_INCREMENT,
    project_id int,
    display_name varchar(250),
    table_name varchar(250),
    metadata text,
    PRIMARY KEY (id,project_id, display_name)
);