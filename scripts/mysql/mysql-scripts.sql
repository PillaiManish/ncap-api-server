# TODO: add columns: created_timestamp, updated_timestamp

CREATE TABLE users (
  id int auto_increment,
  name varchar(250),
  email varchar(250) UNIQUE ,
  password varchar(250),
PRIMARY KEY (id)
);

CREATE TABLE project_details (
    id int AUTO_INCREMENT,
    user_id int,
    name varchar(250),
    description varchar(250),
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE project_metadata (
    id int AUTO_INCREMENT,
    project_id int,
    display_name varchar(250),
    table_name varchar(250),
    description varchar(250),
    metadata text,
    PRIMARY KEY (id,project_id, display_name)
);

-- TODO add role
CREATE TABLE users_projects (
    user_id int,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,

    project_id int,
    FOREIGN KEY (project_id) REFERENCES project_details(id) ON DELETE CASCADE

    role int default 1,
    status int default 0,

    PRIMARY KEY (user_id, project_id)

)
