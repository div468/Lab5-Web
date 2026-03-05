# Laboratorio 5 - Server Side Rendering
Implementación de un servidor HTTP local simulando un tracker de series que soporta interacción verdadera con el usuario. Implementando dos mecanismos web:
- Formularios HTML con POST
- fetch() haciendo uso de JavaScript

## Challenges Implementados
| Challenge | Puntos |
| :--- | :---: | 
|Botón para eliminar un episodio visto | 10 |
| Implementación de una barra de progreso (episodios vistos vs totales) | 15 |
| Actualización de datos sin reload | 20 |
|Ordenamiento por columnas| 20 |
|Paginación| 40 |  

## ⚙️ Requisitos

Tener instalado:
- Go (1.20 o superior)

Verificar con:
```
go version 
```

## 📦 Instalación

Clonar el repositorio:

```
git clone git@github.com:div468/Lab5-Web.git
cd Lab5-Web
```

Instalar dependencias:

```
go mod tidy
```

## ▶️ Ejecutar el programa
```
go run .
```
o también:
```
go run main.go handlers.go
````

El servidor local iniciará en:

```
http://localhost:8080
```

## 🗃️ Base de Datos

El laboratorio utiliza SQLite y el archivo:
```
series.db
```
que se encuentra incluido en el repositorio.

## 🪄 Funcionalidades

- Crear nuevas series
- Incrementar episodios sin necesidad de recargar la página
- Barra de progreso
- Paginación
- Ordenamiento de tabla vía JavaScript

## (●'◡'●) Autor
Julián Divas