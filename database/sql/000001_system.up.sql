create table if not exists users_ (
  id_ integer primary key autoincrement,
  username_ varchar(32) not null unique,
  email_ varchar(256) not null unique,
  verified_ boolean default false
);

create table if not exists public_keys_ (
  id_ integer primary key autoincrement,
  public_key_sha1_ char(128) not null,

  user_id_ integer not null,
  foreign key (user_id_) references users_(id_)
);

create table if not exists audit_ (
  id_ integer primary key autoincrement,
  session_ char(64),
  timestamp_ datetime default current_timestamp,
  action_ varchar(8), -- get/set/list/delete
  status_ varchar(8), -- success/error
  address_ varchar(16), -- ip address
  client_ varchar(64), -- syringe/ssh

  public_key_id_ integer not null,
  user_id_ integer not null,
  foreign key (public_key_id_) references public_keys_(id_),
  foreign key (user_id_) references users_(id_)
);
