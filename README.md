# fffpro_js

**fffpro_js** is a fast, resilient JavaScript fetcher and beautifier built for bug bounty and reconnaissance workflows.

It automatically downloads JavaScript files, beautifies them using `js-beautify`, and saves them in a structured, readable format.

---

## 🚀 Features

* ⚡ High-performance worker pool
* 🔁 Automatic retries with backoff
* 🎯 Filters only JavaScript files
* 🧹 Built-in JS beautification
* 📂 Clean and traceable output
* 🔗 URL mapping stored inside files
* 🧠 Designed for large-scale recon

---

## 📦 Installation

```bash
go install github.com/nvk0x/fffpro_js@latest
```

---

## ⚙️ Requirements

Install js-beautify:

```bash
npm install -g js-beautify
```

---

## 🛠 Usage

```bash
cat urls.txt | fffpro_js
```

---

## ⚡ Example

```bash
cat urls.txt \
| grep "\.js" \
| fffpro_js -w 100 -timeout 30
```

---

## 📂 Output

```
out_js/
 └── domain.com/
      └── static_js_main_xxx.js
```

Each file contains:

```js
// URL: https://target.com/file.js
```

---

## 🔗 Workflow

Best used with httpx:

```bash
cat urls.txt \
| httpx -silent \
| grep "\.js" \
| fffpro_js
```

---

## ⚠️ Disclaimer

For authorized security testing only.

---

## 📜 License

MIT License

