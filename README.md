# Fichenn

Fichenn is a standalone binary - written in Golang - for uploading and downloading secured files.

It aims to be portable and have a straightforward stream implementation (no fancy customisation based on third libraries like cURL and OpenSSL).

## Usage

```
[~]>> finn ~/.fichennrc
Passphrase: 0|l5ZJW#ZlqWFlfW(0l5q#WQ

⠸ uploading (276 B, 0.716 kB/s)
Command:
 finn --pass "0|l5ZJW#ZlqWFlfW(0l5q#WQ" "https://plik.root.gg/file/dfhJOmsP6xOpMnPG/SjJA2l1ZA6kOxizn/.fichennrc" -o ".fichennrc"
Copied to the clipboard
```

```
[~]>> finn ~/.vim
Passphrase: t(gg(tHHRhHy(2(aH2=HuhQH

⠼ uploading (8.1 MB, 1.109 MB/s)
Command:
 finn --pass "t(gg(tHHRhHy(2(aH2=HuhQH" "https://plik.root.gg/file/6HRYSfIZH7uiPAjn/IZYdbDFJe4HtxH9p/.vim.tar" -o ".vim.tar" --extract
Copied to the clipboard
```

## How does it work?

- Upload stream workflow

```
file -> zstd -> stream-chunked-encryption(chacha20poly1305) -> storage-server
```
```
directory -> tarball -> zstd -> stream-chunked-encryption(chacha20poly1305) -> storage-server
```

- Download stream workflow

```
storage-server -> stream-chunked-encryption(chacha20poly1305) -> zstd -> file
```
```
storage-server -> stream-chunked-encryption(chacha20poly1305) -> zstd -> tarball -> directory
```

> - Compressed with Zstandard algorithm
> - Stream chunked encryption scheme with chacha20poly1305 algorithm (borrowed from [age](https://github.com/FiloSottile/age))

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