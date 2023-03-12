#crud: {
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
				filter: ["eq", "ne", "lt", "le", "gt", "ge"]
			}
			modified_at: {
				type:     "string"
				format:   "date-time"
				required: false
				readOnly: true
				filter: ["eq", "ne", "lt", "le", "gt", "ge"]
			}
			timestamp: {
				type:     "string"
				format:   "date-time"
				required: true
				readOnly: false
				filter: ["eq", "ne", "lt", "le", "gt", "ge"]
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
	url:         "http://api.example.com/v1"
	description: "Servidor ejemplo"
}]
paths: {for resource, data in #crud {
	"/api/\(data.path)": {
		get: {
			summary: "Queries a list of \(resource)"
			description: """
				All query (q) parameters support several **operators**:

				- `eq`: equals
				- `ne`: not equals
				- `lt`: less than
				- `le`: less or equal
				- `gt`: greater than
				- `ge`: greater or equal
				- `like`: SQL like
				"""
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
			}
		}
		post: {
			summary: "Creates a new \(resource)"
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
		}
		get: {
			summary:    "Queries a \(resource) by id"
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
			}
		}
		put: {
			summary:    "Updates a \(resource) by id"
			parameters: #param_id
			responses:  #empty_response
		}
		delete: {
			summary:    "Deletes a \(resource) by id"
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
			type: "string"
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
