create table if not exists projects_ (
  id_ integer primary key autoincrement,
  name_ varchar(256) unique not null
);

create table if not exists environments_ (
  id_ integer primary key autoincrement,
  name_ varchar(256) not null,
  project_id_ integer not null,

  foreign key (project_id_) references projects_(id_) on delete cascade
);

create table if not exists secrets_ (
  id_ integer primary key autoincrement,
  key_ text not null unique,
  value_ text not null,
  environment_id_ integer not null,

  foreign key (environment_id_) references environments_(id_) on delete cascade
);

