create table if not exists store_ (
  id_ integer primary key autoincrement,
  key_ text not null unique,
  value_ text not null unique
);

