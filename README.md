# PC Info Gather Tool

This application retrieves system information using Windows Management Instrumentation (WMI) and exports it to a JSON file.

## How to Build

1. Make sure you have Go installed on your system. You can download it from https://golang.org/dl/.
2. Clone this repository to your local machine.
3. Open a terminal and navigate to the directory containing the main.go file.
4. Run the following command to build the application: `go build main.go`

This will generate an executable file `main.exe` 

## How to Run

1. After building the application, you can run it by simple double-clicking the exe
2. The application will retrieve system information, export it to a file named `system_info.json`, and wait for you to press Enter before closing.

## Note

- This application is designed to run on Windows systems.
- Make sure you have the necessary permissions to access WMI information on your system.

## Dependencies

- github.com/yusufpapurcu/wmi
- golang.org/x/sys/windows

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
