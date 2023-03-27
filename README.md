# VideoAPI

## Actualización del swagger

La descripción swagger de la API se genera a partir del fichero [swagger.cue](internal/swagger/swagger.cue). Si se modifican los endpoints, parámetros o tipos de datos de la API, se debe actualizar el fichero `cue` y regenerar la documentación con los comandos:

```
cue fmt internal/swagger/swagger.cue
go generate ./...
```

## Ejecución con docker-compose

Este repositorio incluye un fichero [docker-compose.yaml](docker-compose.yaml) con la especificación adecuada para poder levantar localmente una instancia de esta API, escuchando en el puerto **8080**.

Paran lanzar la instancia, se deben ejecutar estos pasos desde el directorio donde se haya clonado el repositorio:

```
# reconstruir las imágenes docker del servicio
docker-compose build
# Levantar la imagen
docker-compose up
```

La API swagger está disponible en la ruta `/swagger`. Las credenciales de administrador por defecto de esta instancia de desarrollo son:

- usuario: superAdmin
- password: superPassword

Dentro del fichero [docker-compose.yaml](docker-compose.yaml), hay un ejemplo de la cadena de conexión que puede usarse para conectar al servidor oracle con `sqlplus`, si fuese necesario.
