```
                          _
   _____  ___    ____    / \  __  ____    ___    _____
  / ___/ / _ \  / _  \  /  / / / / __ \  / _ \  /  __/
 (__  ) /  __/ / /_/ / /  / / / / / / / /  __/ (__  )
/____/  \____\ \_   /  \_/  \/  \/ /_/  \____\/____/
             ___/  /
             \____/
```

# seglines
Segment unsegmented regions into line snippets.  The regions need to
be created using [segregs](https://github.com/cisocrgoup/segregs).

## Usage
`seglines JSON [JSON...]`

Segment the given regions into lines.  If the given region already
appears to be line segmented, the region is skipped.

## Installation
To install just type `go get github.com/finkf/seglines`.
