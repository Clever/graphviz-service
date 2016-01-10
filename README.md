# Graphviz service

A REST interface to the [graphviz](http://www.graphviz.org/) command line tool.

## Endpoints:

- **/dot?[format=png|svg|pdf|plain]** `POST`: Expects the post body to be in the [dot format](https://en.wikipedia.org/wiki/DOT_(graph_description_language)) and returns the rendered graph is the specified format.  Format defaults to png.
