create table if not exists users_ (
  id_ integer primary key autoincrement,
  username_ varchar(256) not null,
  email_ varchar(256) not null,
  created_at_ datetime without time zone default current_timestamp,
  status_ varchar(8) not null
);

create table if not exists keys_ (
  id_ integer primary key autoincrement,
  ssh_public_key_ varchar(1024) not null,
  user_id_ integer not null,
  created_at_ datetime without time zone default current_timestamp,

  foreign key (user_id_) references users_(id_)
);

