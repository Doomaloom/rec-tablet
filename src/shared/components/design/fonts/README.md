# Rubik font assets

Source: Google Fonts' Rubik family, downloaded October 16, 2026:

- https://github.com/google/fonts/tree/main/ofl/rubik
- `Rubik[wght].ttf` and `Rubik-Italic[wght].ttf`
- `OFL.txt` is included unchanged with its copyright notice and SIL Open Font License.

The bundled static TTF files were instantiated from those variable fonts with
fontTools 4.64.0 at weights 400 (regular) and 600 (semibold), preserving the italic
axis supplied by the italic source. Fyne's bold text style maps to semibold to
match the web collection's `font-semibold`. Monospace text uses Fyne's standard
monospace resource.

No network access is needed when building or running the app.
