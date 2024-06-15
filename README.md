<div align="center">
[![Docker Pulls](https://img.shields.io/docker/pulls/spanglesontoast/arcwiki)](https://hub.docker.com/r/spanglesontoast/arcwiki)
[![License](https://img.shields.io/badge/license-GPLv3-blue.svg?style=flat)](https://github.com/requarks/wiki/blob/master/LICENSE)
# ArcWiki
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://cdn.jsdelivr.net/gh/ArcWiki/ArcWiki@2221be3f4becabe2d61d9da0e9d5114979f7a2be/assets/images/arcwiki.svg">
  <img alt="Wiki.js" src="https://cdn.jsdelivr.net/gh/ArcWiki/ArcWiki@2221be3f4becabe2d61d9da0e9d5114979f7a2be/assets/images/arcwiki.svg" width="100">
</picture>
</div>





## What is it?
A community-driven Go wiki inspired by ArchWiki.

## Usage
Can be used as selfhosted personal wiki for the moment can be used with or without docker. 
Auth is simple and dangerous see admin.json.

## Docker Instructions

``` docker run --name arcwiki -p 8080:8080 -d spanglesontoast/arcwiki ```

or

you can specify the username or password respectively:

``` docker run -e USERNAME=jack -e PASSWORD=pumpkin --name arcwiki -p 8080:8080 -d spanglesontoast/arcwiki ```

you can specify sitename with SITENAME enviroment variable

``` docker run -e SITENAME="Marvel Wiki" -e USERNAME=jack -e PASSWORD=pumpkin --name arcwiki -p 8080:8080 -d spanglesontoast/arcwiki ```
