# Fichenn

Fichenn is a standalone binary - written in Golang - for uploading and downloading secured files.

It aims to be portable and have a straightforward stream implementation (no fancy customisation based on third libraries like cURL and OpenSSL). The drawback is that you need to install this binary on both source and destination operating system.

## Usage

```
[~]>> finn -h
Fichenn secured uploads

Usage:
  finn [flags]

Flags:
  -c, --chmod+x         perform `chmod +x' on downloaded file
  -x, --extract         Tarball extract
  -h, --help            help for finn
  -o, --output string   write output to given destination
  -p, --pass string     passphrase used to decrypt
  -v, --version         version for finn
```

```
[~]>> finn ~/.fichennrc
Passphrase: nF1wCZ8nHv(in|GqaVkWq~iw

⠸ uploading (276 B, 0.716 kB/s)
Command:
 finn --pass "nF1wCZ8nHv(in|GqaVkWq~iw" "https://plik.root.gg/file/dfhJOmsP6xOpMnPG/SjJA2l1ZA6kOxizn/.fichennrc" -o ".fichennrc"
Copied to the clipboard
```

```
[~]>> finn ~/.vim
Passphrase: G8iORZz4pMU=O4i#6cGEJ~ci

⠼ uploading (8.1 MB, 1.109 MB/s)
Command:
 finn --pass "G8iORZz4pMU=O4i#6cGEJ~ci" "https://plik.root.gg/file/6HRYSfIZH7uiPAjn/IZYdbDFJe4HtxH9p/.vim.tar" -o ".vim.tar" --extract
Copied to the clipboard
```

## How does it work?

- Upload stream workflow

```
file -> zstd -> age-encryption.org/v1 -> storage-server
```
```
directory -> tarball -> zstd -> age-encryption.org/v1 -> storage-server
```

- Download stream workflow

```
storage-server -> age-encryption.org/v1 -> zstd -> file
```
```
storage-server -> age-encryption.org/v1 -> zstd -> tarball -> directory
```

> - Compressed with Zstandard algorithm
> - Stream chunked encryption scheme with [age](https://github.com/FiloSottile/age)

## Storages

- [Plik](https://plik.root.gg/) (https://github.com/root-gg/plik)

## License

**MIT**


## Contributing

All PRs are welcome.

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request
