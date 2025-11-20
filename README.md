# 🇩🇪 German Word Translator (Go CLI)

```
 
        German → English Dictionary CLI (Go)

```

A tiny command-line tool that fetches **English translations** for a German word using the **Langenscheidt** dictionary.

---

## Build


```bash
go build german.go
```

This creates an executable.

---

## Usage

Search for a German word:

```bash
./german <german-word>
```
Example:

![example](/example.png)

---

## Notes

* Requires Go installed
* An internet connection
* Only prints up to 5 clean translations (for now)

**Be nice on the server, no heavy requests.**
