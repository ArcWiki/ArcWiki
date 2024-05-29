# ArcWiki
A community-driven Go wiki inspired by ArchWiki.

## Usage
Can be used as selfhosted personal wiki for the moment can be used with or without docker. 
Auth is simple and dangerous see admin.json might change this soon.

## Docker Instructions

$ docker pull spanglesontoast/arcwiki:latest

$ docker run --name bluearcwiki -p 8080:8080 -d spanglesontoast/arcwiki
