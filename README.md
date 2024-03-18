# Impromon

Impromon is for those times when you need to monitor just a few services, but you
don't want to spin up infrastructure.

## Table of Contents
- [Getting Started](#getting-started)
- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## Getting Started

You can download the binary from the releases page or you can build it yourself.

```go build```
## Installation

You can just copy the binary to a location in your path or run it directly from the 
directory where you downloaded it.

## Usage

```./impromon -u https://www.google.com -u https://www.yahoo.com```

or you can generate a file with each service on newlines.
    
```./impromon -s example.lst```

The file will look like the following
    
```
https://www.google.com
https://www.yahoo.com
http://www.example.com
tcp://ad.example.com:636
```

## Contributing

If you would like to contribute, please open an issue or a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

