drop table if exists user;
create table if not exists user (
    id integer primary key,
    name text,
    age integer,
    role varchar(64),
    created integer
);