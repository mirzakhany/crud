insert into users (name, email, password) values ('admin', 'admin@admin.com', 'admin');
insert into users (name, email, password) values ('mohsen', 'mohsen@admin.com', 'admin');
insert into users (name, email, password) values ('mina', 'mina@admin.com', 'admin');

insert into organizations (name) values ('apifor.dev');

insert into permissions (name) values ('users:create');
insert into permissions (name) values ('users:update');
insert into permissions (name) values ('users:list');
insert into permissions (name) values ('users:delete');
insert into permissions (name) values ('organizations:create');
insert into permissions (name) values ('organizations:update');
insert into permissions (name) values ('organizations:list');
insert into permissions (name) values ('organizations:delete');
