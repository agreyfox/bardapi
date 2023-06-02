# AI Bot Unofficial API

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

> A collection of Go code for handling Bing chat bot, Bard chat bot, and Bing Image creator API.

## Table of Contents

- [AI Bot Unofficial API](#ai-bot-unofficial-api)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Features](#features)
  - [Getting Started](#getting-started)
  - [Usage](#usage)
  - [License](#license)
  - [Contributing](#contributing)
  - [Support](#support)

## Overview

This project provides Go code for handling Bing chat bot, Bard chat bot, and Bing Image creator API. It offers functionality to interact with these services and can be used as a starting point for developing chat bots and working with Bing Image creation.

## Features

- Integration with Bing chat bot
- Integration with Bard chat bot
- Utilization of Bing Image creator API

## Getting Started

To get started with the project, follow these steps:

1. Clone the repository:

   ```shell
   git clone https://github.com/your-username/your-repo.git
   ```

2. Install Go and set up your Go environment. Refer to the official Go documentation for instructions.
3. Install the project dependencies:
   ```shell
   go get -u ./...
   ```
4. Configure the necessary API keys and credentials by following the instructions in the configuration file.

## Usage

Here are test examples of how to use the code:

```
func TestNewBingImageClient(t *testing.T) {

	bi, _ := NewImageGen(Cookie_U, false, "test.log")
	img, err := bi.GetImages("Film still of an elderly wise yellow man playing chess, medium shot, mid-shot")
	if err != nil {
		t.Errorf("error %s", err)
	}
	for _, item := range img {
		t.Logf(item)
	}
	bi.MakeThumbnail(true)
	bi.SaveImages(img, ".", "img")
}
```

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, feel free to open an issue or create a pull request.

Please make sure to follow the code of conduct in all your interactions with the project.

## Support

If you have any questions or need assistance, please reach out to e_raeb@yahoo.com.
