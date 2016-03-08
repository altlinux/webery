{
	"version": 1,
	"endpoints": [
		{
			"url": "{schema}://{host}/api/v1/search",
			"method": {
				"GET": {
					"description": "Returns a list of tasks and subtasks",
					"parameters": [
						{
							"type": "string",
							"name": "prefix",
							"description": "Filter objects by prefix",
							"default": "NaN"
						},
						{
							"type": "number",
							"name": "limit",
							"description": "shows only specified number of retults",
							"default": "1000"
						}
					]
				}
			}
		},
		{
			"url": "{schema}://{host}/api/v1/tasks",
			"method": {
				"GET": {
					"description": "Returns a list of tasks",
					"parameters": [
						{
							"type": "string",
							"name": "state",
							"description": "shows tasks with specified state",
							"default": "NaN"
						},
						{
							"type": "string",
							"name": "owner",
							"description": "shows tasks with specified owner",
							"default": "NaN"
						},
						{
							"type": "number",
							"name": "limit",
							"description": "shows only specified number of retults",
							"default": "1000"
						}
					]
				},
				"POST": {
					"description": "Creates new task"
				},
				"DELETE": {
					"description": "Not allowed"
				}
			}
		},
		{
			"url": "{schema}://{host}/api/v1/tasks/{taskid}",
			"method": {
				"GET": {
					"description": "Returns information about specified task"
				},
				"POST": {
					"description": "Not allowed"
				},
				"DELETE": {
					"description": "Removes specified task"
				}
			}
		},
		{
			"url": "{schema}://{host}/api/v1/tasks/{taskid}/subtasks",
			"method": {
				"GET": {
					"description": "Returns information about specified subtask",
					"parameters": [
						{
							"type": "number",
							"name": "limit",
							"description": "shows only specified number of retults",
							"default": "1000"
						}
					]
				},
				"POST": {
					"description": "Not allowed"
				},
				"DELETE": {
					"description": "Not allowed"
				}
			}
		},
		{
			"url": "{schema}://{host}/api/v1/tasks/{taskid}/subtasks/{subtaskid}",
			"method": {
			"GET": {
					"description": "Returns information about specified subtask"
				},
				"POST": {
					"description": "Not allowed"
				},
				"DELETE": {
					"description": "Not allowed"
				}
			}
		},
		{
			"url": "{schema}://{host}/api/v1/acl/{repo}/packages",
			"method": {
				"GET": {
					"description": "Returns list with all packages",
					"parameters": [
						{
							"type": "string",
							"name": "prefix",
							"description": "Filter name by prefix",
							"default": "NaN"
						},
						{
							"type": "string",
							"name": "name",
							"description": "Filter only specified name",
							"default": "NaN"
						},
						{
							"type": "string",
							"name": "member",
							"description": "Filter objects that contains specified member",
							"default": "NaN"
						},
						{
							"type": "number",
							"name": "limit",
							"description": "shows only specified number of retults",
							"default": "1000"
						}
					]
				}
			}
		},
		{
			"url": "{schema}://{host}/api/v1/acl/{repo}/packages/{name}",
			"method": {
				"GET": {
					"description": "Shows ACL for specified package"
				}
			}
		},
		{
			"url": "{schema}://{host}/api/v1/acl/{repo}/groups",
			"method": {
				"GET": {
					"description": "Returns list with all groups",
					"parameters": [
						{
							"type": "string",
							"name": "prefix",
							"description": "Filter name by prefix",
							"default": "NaN"
						},
						{
							"type": "string",
							"name": "name",
							"description": "Filter only specified name",
							"default": "NaN"
						},
						{
							"type": "string",
							"name": "member",
							"description": "Filter objects that contains specified member",
							"default": "NaN"
						},
						{
							"type": "number",
							"name": "limit",
							"description": "shows only specified number of retults",
							"default": "1000"
						}
					]
				}
			}
		},
		{
			"url": "{schema}://{host}/api/v1/acl/{repo}/groups/{name}",
			"method": {
				"GET": {
					"description": "Shows ACL for specified group"
				}
			}
		}
	]
}
