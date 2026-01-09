# LiveKit Load Testing Guide

This guide explains how to perform load testing on your LiveKit server to determine how many concurrent users/rooms it can handle.
We use the official LiveKit CLI (`lk`) which includes a `load-test` command to simulate multiple publishers and subscribers.

## 1. Install LiveKit CLI (`lk`)

Since you are on Windows and `scoop` is not available, you need to install it manually:

1.  **Download**: Go to the [LiveKit CLI Configuration](https://github.com/livekit/livekit-cli/releases) page.
2.  **Select Version**: Click on the latest version (e.g., `v1.x.x`).
3.  **Asset**: Download `livekit-cli_1.x.x_windows_amd64.tar.gz` (or `.zip` if available).
4.  **Extract**: Extract the `lk.exe` file to a folder (e.g., `C:\lk\`).
5.  **Path (Optional but recommended)**: Add that folder to your System PATH so you can run `lk` from any terminal.
    -   *Search "Edit the system environment variables" in Windows Search -> Environment Variables -> Path -> Edit -> New -> Paste path to folder.*

## 2. Verify Installation

Open a **new** terminal (PowerShell or Command Prompt) and run:

```powershell
lk version
```

If it prints a version number, you are ready.

## 3. Running a Load Test

The `load-test` command simulates users joining a room.

### Basic Connectivity Test
Run this to ensure everything works (1 publisher, 1 subscriber):

```powershell
lk load-test --url <your_livekit_url> --api-key <your_api_key> --api-secret <your_api_secret> --video-publishers 1 --subscribers 1 --duration 30s
```

*Replace `<...>` with your actual values from your `.env` file.*

### Concurrency Test (Finding the Limit)
To find the limit, progressively increase the number of publishers and subscribers.

**Example: Simulate 50 users (5 video publishers, 45 subscribers)**

```powershell
lk load-test --url <your_livekit_url> --api-key <your_api_key> --api-secret <your_api_secret> --video-publishers 5 --subscribers 45 --duration 1m
```

### Testing a Specific Room
To test a specific room (e.g., one that already exists or to group testers), use the `--room` flag:

```powershell
lk load-test --room <room_name> --url <...> ...
```

### Important Flags

-   `--url`: URL of your LiveKit server (e.g., `ws://localhost:7880` or `wss://your-cloud-project.livekit.cloud`).
-   `--api-key`, `--api-secret`: Your LiveKit credentials.
-   `--video-publishers`: Number of users publishing video.
-   `--audio-publishers`: Number of users publishing audio.
-   `--subscribers`: Number of users who only watch/listen.
-   `--duration`: How long the test runs (e.g., `1m`, `30s`).
-   `--room`: (Optional) Specify a room name. If omitted, a random room is created.

## 4. Interpreting Results

The tool will output stats like:
-   **Bitrate**: Total bandwidth used.
-   **Packet Loss**: Should be < 1% for good quality.
-   **Latency/RTT**: Should be low (e.g., < 200ms).
-   **Drop Rate**: If this increases, your server is reaching its limit.

If you see high packet loss or connection errors, you have hit the concurrency limit for your current server resources.
