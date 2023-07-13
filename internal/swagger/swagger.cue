#crud: {

	User: {
		path:      "user"
		mediaType: ""
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
				required: true
				readOnly: false
				filter: ["eq", "ne", "like"]
			}
			role: {
				type:     "string"
				required: true
				readOnly: false
				filter: []
				enum: ["ADMIN", "READ_ONLY", "READ_WRITE", "SERVICE"]
			}
			password: {
				type:     "string"
				required: true
				readOnly: false
				filter: []
			}
		}
	}

	Camera: {
		path:      "camera"
		mediaType: ""
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
				required: true
				readOnly: false
				filter: ["eq", "ne", "like"]
			}
			latitude: {
				type:     "number"
				required: true
				readOnly: false
				filter: ["lt", "le", "gt", "ge"]
			}
			longitude: {
				type:     "number"
				required: true
				readOnly: false
				filter: ["lt", "le", "gt", "ge"]
			}
			local_path: {
				type:     "string"
				required: false
				readOnly: false
				filter: []
			}
		}
	}

	Video: {
		path:      "video"
		mediaType: "video/4gpp, video/3gpp2, video/3gp2, video/mpeg, video/mp4, video/ogg, video/quicktime, video/webm"
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
			media_url: {
				type:     "string"
				required: false
				readOnly: true
				filter: ["eq", "ne"]
			}
		}
	}

	Picture: {
		path:      "picture"
		mediaType: "image/jpeg, image/png"
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
			media_url: {
				type:     "string"
				required: false
				readOnly: true
				filter: ["eq", "ne"]
			}
		}
	}

	Alert: {
		path:      "alert"
		mediaType: ""
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
			severity: {
				type:     "string"
				required: true
				readOnly: false
				filter: ["eq", "ne", "like"]
			}
			message: {
				type:     "string"
				required: true
				readOnly: false
				filter: ["eq", "ne", "like"]
			}
			acknowledged_at: {
				type:     "string"
				format:   "date-time"
				required: false
				readOnly: false
				filter: ["lt", "le", "gt", "ge", "eq", "ne"]
			}
			resolved_at: {
				type:     "string"
				format:   "date-time"
				required: false
				readOnly: false
				filter: ["lt", "le", "gt", "ge", "eq", "ne"]
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

// DataTypes
// ---------
components: schemas: queryError: {
	type: "object"
	properties: error: type: "string"
}

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
			prev: {
				type:    "string"
				example: "sort=asc&offset=0&limit=10"
			}
			data: {
				type: "array"
				items: {
					"$ref": "#/components/schemas/\(resource)"
				}
			}
		}
	}
	"\(resource)": {
		type: "object"
		properties: {for propname, propdata in data.properties {
			(propname): {
				type: propdata.type
				if (type == "array") {
					items: type: "string"
				}
				if propdata.enum != _|_ {
					enum: propdata.enum
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
	"put_\(resource)": {
		type: "object"
		properties: {for propname, propdata in data.properties if propname != "id" && !propdata.readOnly {
			(propname): {
				type: propdata.type
				if (type == "array") {
					items: type: "string"
				}
				if propdata.enum != _|_ {
					enum: propdata.enum
				}
				if propdata.format != _|_ {
					format: propdata.format
				}
				if propdata.readOnly != _|_ {
					readOnly: propdata.readOnly
				}
			}
		}}
		required: false
	}
}}

// Security schemas
// ----------------
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

components: securitySchemes: apiKey: {
	type: "apiKey"
	"in": "query"
	name: "apiKey"
}

// Abbreviations
// -------------
#queryErrorReference: {
	"application/json": schema: "$ref": "#/components/schemas/queryError"
}

#standardResponses: {
	"401": {
		description: "Unauthorized"
	}
	"400": {
		description: "Invalid query"
		content:     #queryErrorReference
	}
	"500": {
		description: "Internal error"
		content:     #queryErrorReference
	}
	...
}

#loginResponses: {
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
	"400": {
		description: "Invalid query"
		content:     #queryErrorReference
	}
	"500": {
		description: "Internal error"
		content:     #queryErrorReference
	}
	...
}

#webhookResponses: {
	"200": {
		description: "Webhook processed successfully"
	}
	"400": {
		description: "Invalid request"
		content: {
			"text/plain": {
				schema: {
					type: "string"
				}
			}
		}
	}
	"500": {
		description: "Internal error"
		content: {
			"text/plain": {
				schema: {
					type: "string"
				}
			}
		}
	}
	...
}

#secured: {
	security: [
		{bearerAuth: []},
		{cookieaAuth: []},
	]
	...
}

// Authentication endpoints
// ------------------------
paths: "/v1/api/login": {
	post: {
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
		responses: #loginResponses
	}
	get: {
		summary: "Refresh the authentication token"
		#secured
		tags: ["Auth"]
		responses: #loginResponses
		responses: "401": {
			description: "Unauthorized"
		}
	}
}

paths: "/v1/api/logout": get: {
	summary: "Removes session cookie"
	#secured
	tags: ["Auth"]
	responses: {
		"204": {
			description: "No data"
		}
		"400": {
			description: "Invalid query"
			content:     #queryErrorReference
		}
		"500": {
			description: "Internal error"
			content:     #queryErrorReference
		}
	}
}

paths: "/v1/api/me": get: {
	summary: "Returns information about the logged-in user"
	#secured
	tags: ["Auth"]
	responses: #standardResponses
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
								enum: ["ADMIN", "READ_ONLY", "READ_WRITE", "SERVICE"]
							}
						}
					}
				}
			}
		}
	}
}

// CRUD endpoints
paths: {for resource, data in #crud {
	"/v1/api/\(data.path)": {
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
						"in":     "query"
						required: false
						_repeatable: bool | *false
						if op == "eq" {
							description: "Find items where field `\(propname)` is `equal` to this value (use `NULL` to match null values)"
							_repeatable: true
						}
						if op == "ne" {
							description: "Find items where field `\(propname)` is `not equal` to this value (use `NULL` to match null values)"
							_repeatable: true
						}
						if op == "gt" {
							description: "Find items where field `\(propname)` is `greater than` this value"
							_repeatable: false
						}
						if op == "ge" {
							description: "Find items where field `\(propname)` is `greater or equal` than this value"
							_repeatable: false
						}
						if op == "lt" {
							description: "Find items where field `\(propname)` is `less than` this value"
							_repeatable: false
						}
						if op == "le" {
							description: "Find items where field `\(propname)` is `less or equal` than this value"
							_repeatable: false
						}
						if op == "like" {
							description: "Find items where field `\(propname)` is `like` to this value"
							_repeatable: true
						}
						schema: {
							if _repeatable {
								type: "array"
								items: {
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
							}
							if !_repeatable {
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
			// I don't want to hide the complexity of a moving count
			// under a single api call. This parameter will be undocumented.
			//#parameters: count: {
			//	"in":        "query"
			//	required:    false
			//	description: "Include total item count in result"
			//	schema: {
			//		type: "boolean"
			//	}
			//}
			parameters: [ for paramname, paramdata in #parameters {
				name: paramname
				paramdata
			}]
			responses: #standardResponses
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
			responses: #standardResponses
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
			}
		}
	}
	"/v1/api/\(data.path)/{id}": {
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
			#standardResponses
		}
		get: {
			summary: "Queries a \(resource) by id"
			tags: [resource]
			#secured
			parameters: #param_id
			responses:  #standardResponses
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
			}
		}
		if data.mediaType != "" {
			post: {
				summary: "Uploads the file for the \(resource) by id"
				tags: [resource]
				#secured
				parameters: [{
					name:     "id"
					"in":     "path"
					required: true
					schema: type: "string"
				}, {
					name:     "redirectOnSuccess"
					"in":     "query"
					required: false
					schema: type: "string"
					description: "If provided, redirect URL on success"
				}, {
					name:     "redirectOnError"
					"in":     "query"
					required: false
					schema: type: "string"
					description: "If provided, redirect URL on error. \"error\" will be appended to queryString"
				}]
				requestBody: content: "multipart/form-data": {
					schema: {
						type: "object"
						properties: file: {
							type:   "string"
							format: "binary"
						}
					}
					encoding: file: contentType: data.mediaType
				}
				responses: #standardResponses
				responses: {
					"200": {
						description: "Media URL for the file uploaded"
						content: "application/json": schema: {
							type: "object"
							properties: id: type:        "string"
							properties: media_url: type: "string"
						}
					}
					"301": {
						description: "Redirect to the provided URLs on success or error"
						headers: Location: {
							description: "URL to redirect to"
							schema: type: "string"
						}
					}
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
						schema: "$ref": "#/components/schemas/put_\(resource)"
					}
				}
			}
			responses: #empty_response
		}
		delete: {
			summary: "Deletes a \(resource) by id"
			tags: [resource]
			#secured
			if data.mediaType == "" {
				parameters: #param_id
			}
			if data.mediaType != "" {
				parameters: [{
					name:     "id"
					"in":     "path"
					required: true
					schema: type: "string"
				}, {
					name:     "mediaOnly"
					"in":     "query"
					required: false
					schema: type: "boolean"
					description: "If true, only the media will be deleted"
				}]
			}
			responses: #empty_response
		}
	}
}}

// Alertmanager webhook
// --------------------
paths: "/v1/api/hook": {
post: {
	summary: "webhook receiver for alertmanager"
	security: [{apiKey: []}]
	tags: ["AlertManager"]
	requestBody: {
		required:    true
		description: "alert body"
		content: {
			"application/json": {
				schema: "$ref": "#/components/schemas/alertmanager_webhook"
			}
		}
	}
	responses: #webhookResponses
}
}

components: schemas: alertmanager_webhook: {
type: "object"
properties: {
	version: type:           "string"
	groupKey: type:          "string"
	status: type:            "string"
	receiver: type:          "string"
	groupLabels: type:       "object"
	commonLabels: type:      "object"
	commonAnnotations: type: "object"
	externalURL: type:       "string"
	alerts: {
		type: "array"
		items: {
			type: "object"
			properties: {
				status: type:      "string"
				labels: type:      "object"
				annotations: type: "object"
				startsAt: {
					type:   "string"
					format: "date-time"
				}
				endsAt: {
					type:   "string"
					format: "date-time"
				}
				generatorURL: type: "string"
			}
		}
	}
}
}
