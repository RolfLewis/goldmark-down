# goldmark-down

A markdown parser for use with [@yuin/goldmark](https://github.com/yuin/goldmark). Whereas [goldmark](https://github.com/yuin/goldmark)'s parsing is mostly focused on Commonmark compatibility, this renderer is a baseline Markdown implementation upon which customizations can be added for arbitrary output.

Supports, by default, Github Markdown Tables, Tasks, and Strikethrough elements. *This functionality will likely be exported to an extension in the future.*

> A significant portion of the contained rendering code is referenced directly from [@pgavlin/goldmark](https://github.com/pgavlin/goldmark), an as-of-writing-unmaintained fork of the original goldmark project that implements a customized markdown renderer. Modifications and improvements have been made on the referenced code, but I extend a special thanks to [@pgavlin](https://github.com/pgavlin/goldmark) for a working implementation.
