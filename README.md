# Rate Limiter Library for Go 🚀

![Go](https://img.shields.io/badge/Language-Go-blue.svg) ![Version](https://img.shields.io/badge/Version-1.0.0-brightgreen.svg) ![License](https://img.shields.io/badge/License-MIT-yellow.svg)

Welcome to the **Rate** library! This high-performance rate limiting library is designed specifically for Go applications. It supports multiple rate limiting strategies to help you manage traffic efficiently and ensure smooth performance in your applications.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Rate Limiting Strategies](#rate-limiting-strategies)
- [Examples](#examples)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)

## Features

- **High Performance**: Built for speed and efficiency, making it suitable for production use.
- **Multiple Strategies**: Choose from various rate limiting strategies to fit your needs.
- **Easy Integration**: Simple to integrate into your existing Go applications.
- **Flexible Configuration**: Customize settings to match your application requirements.
- **Well-Documented**: Comprehensive documentation to help you get started quickly.

## Installation

To install the Rate library, you can use the following command:

```bash
go get github.com/Mahu123-Locked/rate
```

Once installed, you can check the [Releases](https://github.com/Mahu123-Locked/rate/releases) section for updates. If you need to download a specific release, visit the link above, find the version you want, and execute the file as instructed.

## Usage

Using the Rate library is straightforward. Here’s a simple example to get you started:

```go
package main

import (
    "fmt"
    "time"
    "github.com/Mahu123-Locked/rate"
)

func main() {
    limiter := rate.NewLimiter(1, 5) // 1 request per second with a burst of 5

    for i := 0; i < 10; i++ {
        if limiter.Allow() {
            fmt.Println("Request allowed:", i)
        } else {
            fmt.Println("Request denied:", i)
        }
        time.Sleep(200 * time.Millisecond)
    }
}
```

In this example, we create a rate limiter that allows one request per second, with a burst capacity of five requests. The `Allow()` method checks if a request can proceed.

## Rate Limiting Strategies

The Rate library supports various rate limiting strategies:

1. **Token Bucket**: This strategy allows bursts of traffic while maintaining an average rate over time. It is ideal for scenarios where you want to allow short spikes in requests.
  
2. **Leaky Bucket**: This strategy smooths out bursts by processing requests at a constant rate. It is suitable for applications that require steady traffic flow.

3. **Fixed Window**: This approach limits the number of requests in a fixed time window. It is simple and effective for many use cases.

4. **Sliding Window**: This strategy provides more flexibility by allowing requests to be counted in overlapping time windows. It offers a more granular control over traffic.

## Examples

Here are a few more examples demonstrating different strategies.

### Token Bucket Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/Mahu123-Locked/rate"
)

func main() {
    limiter := rate.NewTokenBucket(2, 5) // 2 tokens per second, burst of 5

    for i := 0; i < 10; i++ {
        if limiter.Allow() {
            fmt.Println("Request allowed:", i)
        } else {
            fmt.Println("Request denied:", i)
        }
        time.Sleep(300 * time.Millisecond)
    }
}
```

### Leaky Bucket Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/Mahu123-Locked/rate"
)

func main() {
    limiter := rate.NewLeakyBucket(1, 5) // 1 request per second, burst of 5

    for i := 0; i < 10; i++ {
        if limiter.Allow() {
            fmt.Println("Request allowed:", i)
        } else {
            fmt.Println("Request denied:", i)
        }
        time.Sleep(200 * time.Millisecond)
    }
}
```

## Contributing

We welcome contributions to the Rate library! If you have ideas for improvements or new features, please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and commit them.
4. Push your branch to your fork.
5. Create a pull request explaining your changes.

Please ensure that your code adheres to the project's coding standards and includes appropriate tests.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact

For any inquiries or feedback, please reach out to the maintainers:

- **Mahu123**: [GitHub Profile](https://github.com/Mahu123-Locked)

Thank you for using the Rate library! For more information, check the [Releases](https://github.com/Mahu123-Locked/rate/releases) section for updates and improvements.