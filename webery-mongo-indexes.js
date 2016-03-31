db["tasks"       ].ensureIndex({taskid: 1},               {background: true, unique: true,  sparse: false});
db["tasks"       ].ensureIndex({"search.key": 1},         {background: true, unique: false, sparse: true });
db["subtasks"    ].ensureIndex({taskid: 1, subtaskid: 1}, {background: true, unique: true,  sparse: false});
db["subtasks"    ].ensureIndex({"search.key": 1},         {background: true, unique: false, sparse: true });
db["acl_packages"].ensureIndex({repo: 1, name: 1},        {background: true, unique: true,  sparse: false});
db["acl_groups"  ].ensureIndex({repo: 1, name: 1},        {background: true, unique: true,  sparse: false});
