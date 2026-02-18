# pokesprite

Inspired by [msikma/pokesprite](https://github.com/msikma/pokesprite) and forked from [pokedextracker/pokesprite](https://github.com/pokedextracker/pokesprite), this repo
is a variation of the same ideas that generates a spritesheet and stylesheet that are
meant to be used for [Pokedex0](https://pokedex0.com).

It currently consists of 7 scripts:

- `rename` - This renames icons from
  [msikma/pokesprite](https://github.com/msikma/pokesprite) to names that can be
  used by the other scripts. Only use this one if you're copying sprites from
  that repo. Read the comment at the top of the file for more info.
- `chop` - This takes in a JSON file explaining the details of an existing
  spritesheet, and it chops it up into individual images. It can also take an SCSS
  file to extract sprites from a pokesprite-style spritesheet.
- `scale` - This takes any images in the `images` directory that are greater
  than the threshold (default 100px) in either dimension (height or width) and
  either scales it by the provided factor (default 0.5) or to the set dimensions
  passed in. This script will modify the images in place.
- `trim` - This takes all images in the `images` directory and trims any excess
  transparency from it. This is so that we can center the sprites based on
  content (non-transparent pixels) and control the padding through CSS.
- `spritesheet` - This takes all the images in the `images` directory and
  stitches them together into a single image.
- `scss` - This uses the images in the `images` directory to generate a `.scss`
  file that lists classes with the correct positions so the spritesheet can be
  used.
- `positions` - This reads the generated `pokesprite.scss` file and generates a
  TypeScript file (`sprite-positions.ts`) containing a map of sprite positions
  with width, height, and background-position data for use in web applications.


To run any of them, it's a simple `task` command:

```sh
task rename
task chop -- data.json
task scale
task trim
task spritesheet
task scss
task positions
```

There's also a `build` task that runs `trim`, `spritesheet`, `scss`, and `positions` in sequence:

```sh
task build
```

## Setup

### Task

Instead of `make`, this project uses [`task`](https://taskfile.dev/#/). It seems
to be a bit cleaner for some specific things that we want to do.

You can find instructions on how to install it
[here](https://taskfile.dev/#/installation).

### Go

To have everything working as expected, you need to have a module-aware version
of Go installed (v1.11 or greater) and `pngcrush`.

To install Go, you can do it any way you prefer. We recommend using
[`goenv`](https://github.com/syndbg/goenv) so that you can use the correct
version of Go for different projects depending on `.go-version` files. In its
current state, the v2 beta of `goenv` can't be installed through `brew`
normally, so you need to fetch from `HEAD` using the following command:

```sh
brew install --HEAD goenv
```

**Note**: If you already have a v1 version of `goenv` installed, you need to
uninstall it first.

Once installed, you can go into this projects directory and run the following to
install the correct version of Go:

```sh
goenv install
```

### `pngcrush`

`pngcrush` is required for the `spritesheet` command. To install it, you can
just run the following command:

```sh
task setup
```
