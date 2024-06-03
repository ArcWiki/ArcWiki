# ArcWiki
A community-driven Go wiki inspired by ArchWiki.

## Usage
Can be used as selfhosted personal wiki for the moment can be used with or without docker. 
Auth is simple and dangerous see admin.json might change this soon.

## Docker Instructions

``` docker run --name arcwiki -p 8080:8080 -d spanglesontoast/arcwiki ```

or

you can specify the username or password respectively:

``` docker run -e USERNAME=jack -e PASSWORD=pumpkin --name arcwiki -p 8080:8080 -d spanglesontoast/arcwiki ```
you can specify sitename with SITENAME enviroment variable
``` docker run -e SITENAME="Marvel Wiki" -e USERNAME=jack -e PASSWORD=pumpkin --name arcwiki -p 8080:8080 -d spanglesontoast/arcwiki ```
