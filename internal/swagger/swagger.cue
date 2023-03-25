#crud: {

	User: {
		path: "user"
		properties: {
			id: {
				type:     "string"
				required: true
				readOnly: false
				filter: ["eq", "ne", "like"]
			}
			created_at: {
				type:     "string"
				format:   "date-time"
				required: false
				readOnly: true
				filter: ["lt", "le", "gt", "ge"]
			}
			modified_at: {
				type:     "string"
				format:   "date-time"
				required: false
				readOnly: true
				filter: ["lt", "le", "gt", "ge"]
			}
			name: {
				type:     "string"
				required: false
				readOnly: false
				filter: ["eq", "ne", "like"]
			}
			role: {
				type:     "string"
				required: false
				readOnly: false
				filter: []
			}
			password: {
				type:     "string"
				required: false
				readOnly: false
				filter: []
			}
		}
	}

	Video: {
		path: "video"
		properties: {
			id: {
				type:     "string"
				required: true
				readOnly: false
				filter: ["eq", "ne", "like"]
			}
			created_at: {
				type:     "string"
				format:   "date-time"
				required: false
				readOnly: true
				filter: ["lt", "le", "gt", "ge"]
			}
			modified_at: {
				type:     "string"
				format:   "date-time"
				required: false
				readOnly: true
				filter: ["lt", "le", "gt", "ge"]
			}
			timestamp: {
				type:     "string"
				format:   "date-time"
				required: true
				readOnly: false
				filter: ["lt", "le", "gt", "ge"]
			}
			camera: {
				type:     "string"
				required: true
				readOnly: false
				filter: ["eq", "ne", "like"]
			}
			tags: {
				type:     "array"
				required: false
				readOnly: false
				filter: ["eq", "ne", "like"]
			}
		}
	}

	Picture: {
		path: "picture"
		properties: {
			id: {
				type:     "string"
				required: true
				readOnly: false
				filter: ["eq", "ne", "like"]
			}
			created_at: {
				type:     "string"
				format:   "date-time"
				required: false
				readOnly: true
				filter: ["lt", "le", "gt", "ge"]
			}
			modified_at: {
				type:     "string"
				format:   "date-time"
				required: false
				readOnly: true
				filter: ["lt", "le", "gt", "ge"]
			}
			timestamp: {
				type:     "string"
				format:   "date-time"
				required: true
				readOnly: false
				filter: ["lt", "le", "gt", "ge"]
			}
			camera: {
				type:     "string"
				required: true
				readOnly: false
				filter: ["eq", "ne", "like"]
			}
			tags: {
				type:     "array"
				required: false
				readOnly: false
				filter: ["eq", "ne", "like"]
			}
		}
	}
}

openapi: "3.0.0"
info: {
	title:       "VideoAPI"
	description: "API para la comunicación con el backend de video"
	version:     "0.1.0"
}
servers: [{
	url:         "/"
	description: "current host"
}]

// Authentication endpoints
paths: "/api/login": post: {
	summary: "Logs in and returns the authentication token"
	security: []
	tags: ["Auth"]
	requestBody: {
		required:    true
		description: "Credentials"
		content: {
			"application/json": {
				schema: {
					type: "object"
					properties: {
						id: {
							type: "string"
						}
						password: {
							type: "string"
						}
					}
				}
			}
		}
	}
	responses: {
		"200": {
			description: "Authentication token"
			content: {
				"application/json": {
					schema: {
						type: "object"
						properties: {
							id: {
								type: "string"
							}
							name: {
								type: "string"
							}
							role: {
								type: "string"
							}
							token: {
								type: "string"
							}
						}
					}
				}
			}
			headers: {
				"Set-Cookie": {
					description: "Authentication cookie"
					schema: {
						type: "string"
					}
				}
			}
		}
		"401": {
			description: "Invalid credentials"
		}
	}
}

paths: "/api/logout": get: {
	summary: "Removes session cookie"
	#secured
	tags: ["Auth"]
	responses: {
		"204": {
			description: "No data"
		}
	}
}

paths: "/api/me": get: {
	summary: "Returns information about the logged-in user"
	#secured
	tags: ["Auth"]
	responses: {
		"200": {
			description: "Authentication token"
			content: {
				"application/json": {
					schema: {
						type: "object"
						properties: {
							id: {
								type: "string"
							}
							name: {
								type: "string"
							}
							role: {
								type: "string"
							}
						}
					}
				}
			}
		}
		"401": {
			description: "Unauthorized"
		}
	}
}

components: securitySchemes: bearerAuth: {
	type:         "http"
	scheme:       "bearer"
	bearerFormat: "JWT"
}

components: securitySchemes: cookieAuth: {
	type: "apiKey"
	"in": "cookie"
	name: "VIDEOAPI_SESSION"
}

// Secured methods use both auth schemes
#secured: security: [
	{bearerAuth: []},
	{cookieaAuth: []},
]

// CRUD endpoints
paths: {for resource, data in #crud {
	"/api/\(data.path)": {
		get: {
			summary: "Queries a list of \(resource)"
			tags: [resource]
			description: """
				All query (q) parameters support several **operators**:

				- `eq`: equals
				- `ne`: not equals
				- `lt`: less than
				- `le`: less or equal
				- `gt`: greater than
				- `ge`: greater or equal
				- `like`: SQL like

				Operators `eq` and `ne` support the special value `NULL` to match
				non-null values in the DB.
				"""
			#secured
			#parameters: {for propname, propdata in data.properties if propdata.filter != _|_ {
				for op in propdata.filter {
					("q:\(propname):\(op)"): {
						"in":        "query"
						required:    false
						description: "Filter field `\(propname)` with the specified operator (`\(op)`)"
						schema: {
							if propdata.format != _|_ {
								format: propdata.format
							}
							if propdata.type == "array" {
								type: "string"
							}
							if propdata.type != "array" {
								type: propdata.type
							}
						}
					}}
			}}
			#parameters: sort: {
				"in":        "query"
				required:    false
				description: "List of columns to sort by"
				schema: {
					type: "array"
					items: {
						type: "string"
					}
				}
			}
			#parameters: ascending: {
				"in":        "query"
				required:    false
				description: "Sort ascending"
				schema: {
					type: "boolean"
				}
			}
			#parameters: offset: {
				"in":        "query"
				required:    false
				description: "Offset for pagination"
				schema: {
					type: "integer"
				}
			}
			#parameters: limit: {
				"in":        "query"
				required:    false
				description: "Limit for pagination"
				schema: {
					type: "integer"
				}
			}
			parameters: [ for paramname, paramdata in #parameters {
				name: paramname
				paramdata
			}]
			responses: {
				"200": {
					description: "List of items"
					content: {
						"application/json": {
							schema:
								"$ref": "#/components/schemas/ListOf\(resource)"
						}
					}
				}
				"401": {
					description: "Unauthorized"
				}
			}
		}
		post: {
			summary: "Creates a new \(resource)"
			tags: [resource]
			#secured
			requestBody: {
				description: "Information of the \(resource)"
				required:    true
				content: {
					"application/json": {
						schema: "$ref": "#/components/schemas/\(resource)"
					}
				}
			}
			responses: {
				"200": {
					description: "New resource created"
					content: {
						"application/json": {
							schema:
								"$ref": "#/components/schemas/ResourceId"
						}
					}
				}
				"401": {
					description: "Unauthorized"
				}
			}
		}
	}
	"/api/\(data.path)/{id}": {
		#param_id: [{
			name:     "id"
			"in":     "path"
			required: true
			schema: type: "string"
		}]
		#empty_response: {
			"204": {
				description: "no content returned if success"
			}
			"401": {
				description: "Unauthorized"
			}
		}
		get: {
			summary: "Queries a \(resource) by id"
			tags: [resource]
			#secured
			parameters: #param_id
			responses: {
				"200": {
					description: "resource content"
					content: {
						"application/json": {
							schema:
								"$ref": "#/components/schemas/\(resource)"
						}
					}
				}
				"401": {
					description: "Unauthorized"
				}
			}
		}
		put: {
			summary: "Updates a \(resource) by id"
			tags: [resource]
			#secured
			parameters: #param_id
			requestBody: {
				description: "Information of the \(resource)"
				required:    true
				content: {
					"application/json": {
						schema: "$ref": "#/components/schemas/\(resource)"
					}
				}
			}
			responses: #empty_response
		}
		delete: {
			summary: "Deletes a \(resource) by id"
			tags: [resource]
			#secured
			parameters: #param_id
			responses:  #empty_response
		}
	}
}}

components: schemas: ResourceId: {
type: "object"
properties: id: type: "string"
}
components: schemas: {for resource, data in #crud {
"ListOf\(resource)": {
	type: "object"
	properties: {
		next: {
			type:    "string"
			example: "sort=asc&offset=10&limit=10"
		}
		data: {
			type: "array"
			items: {
				"$ref": "#/components/schemas/\(resource)"
			}
		}
	}
}
(resource): {
	type: "object"
	properties: {for propname, propdata in data.properties {
		(propname): {
			type: propdata.type
			if (type == "array") {
				items: type: "string"
			}
			if propdata.format != _|_ {
				format: propdata.format
			}
			if propdata.readOnly != _|_ {
				readOnly: propdata.readOnly
			}
		}
	}}
	required: [ for propname, propdata in data.properties
		if propdata.required {propname}]
}
}}
