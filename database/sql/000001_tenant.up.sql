create table if not exists store_ (
  id_ integer primary key autoincrement,
  key_ varchar(255) not null unique,
  value_ varchar(2048) not null unique
);

