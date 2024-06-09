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