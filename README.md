````markdown
# GoTorrent

## About The Project

GoTorrent is a simplified, peer-to-peer (P2P) file-sharing application built in Go. It allows users to create torrent-like files, seed (share) content, and download files in chunks from peers. The current implementation uses direct TCP connections for file transfer, providing a foundational understanding of how such systems can operate.

### Features

* **Torrent File Creation**: Generate `.torrent` files that contain metadata about the shared file, including its name, size, chunk size, and SHA256 hashes of each chunk for integrity verification.
* **File Seeding**: Act as a seeder to serve file chunks to downloaders upon request.
* **File Downloading**: Download files by requesting chunks from available peers and reconstructing the original file.
* **Chunk-based Transfer**: Files are divided into smaller chunks, enabling parallel downloads and integrity checking.
* **Progress Tracking**: Basic statistics and progress reporting for both seeding and downloading.
* **Hash Verification**: Ensures data integrity by verifying the SHA256 hash of each downloaded chunk against the `.torrent` file's metadata.

### Built With

* [Go](https://go.dev/)

## Getting Started

To get a local copy up and running, follow these simple steps.

### Prerequisites

* Go installed on your system.

### Installation

1.  Clone the repository (or copy the provided code into files):
    ```bash
    git clone <repository_url> # Replace with your repo URL if applicable
    cd gotorrent
    ```
2.  Build the application:
    ```bash
    go build -o gotorrent .
    ```

## Usage

GoTorrent can be run in three modes: `create`, `seed`, or `download`.

### 1. Create a Torrent File

This mode generates a `.torrent` file for a specified local file.

```bash
go run . create <filepath_to_share>
# Example: go run . create my_large_file.txt
````

This will create a `my_large_file.txt.torrent` file in the `./torrents` directory.

### 2\. Start Seeding

This mode starts a seeder that listens for incoming download requests for a specified file. The seeder will make the `filepath` specified during the create process available.

```bash
go run . seed <filepath_of_original_file>
# Example: go run . seed my_large_file.txt
```

The seeder will listen on `127.0.0.1:8080`.

### 3\. Start Downloading

This mode initiates the download of a file using a `.torrent` file.

```bash
go run . download <path_to_torrent_file>
# Example: go run . download ./torrents/my_large_file.txt.torrent
```

The downloaded file will be saved in the `./downloads` directory.

## Future Requirements & Scalability

The current implementation provides a basic foundation. To evolve into a publicly accessible and scalable file-sharing system, the following enhancements are envisioned:

  * **Transition to HTTP for Chunk Serving**:
      * **Echo Framework Integration**: Replace the direct TCP listener in the seeder with an [Echo](https://echo.labstack.com/) web framework. This will allow seeders to expose HTTP endpoints (e.g., `/chunks/:fileName/:chunkIndex`) for serving file chunks. This leverages standard web protocols, making it easier to deploy on cloud platforms and integrate with existing web infrastructure.
      * **HTTP Client for Downloader**: Modify the downloader to use standard HTTP requests to fetch chunks from seeders, moving away from raw TCP connections for data transfer.
  * **Publicly Hosted Backend**:
      * **Cloud Platform Deployment**: Deploy the seeder and potential tracker components on free backend servers like [Railway](https://railway.app/) or similar Platform as a Service (PaaS) providers. This requires careful consideration of how these platforms expose custom TCP/HTTP ports.
      * **Dynamic IP Handling**: Implement Dynamic DNS (DDNS) if seeder nodes have dynamic IP addresses, ensuring that their public endpoints remain resolvable.
  * **Centralized Tracker/Discovery Service**:
      * **HTTP-based Tracker**: Introduce a separate, centralized HTTP service (also built with Echo) that acts as a tracker. Seeders would register the files they are sharing and their public addresses with this tracker. Downloaders would query the tracker to get a list of available peers for a desired file. This decouples peer discovery from hardcoded peer lists.
  * **API Documentation with Swagger/OpenAPI**:
      * **Automated Documentation**: Utilize tools like `swag` or `go-swagger` to generate interactive API documentation for the HTTP endpoints (e.g., for chunk serving and tracker services). This improves maintainability and allows for easier client integration.
  * **Enhanced Scalability & Resiliency**:
      * **Load Balancing**: With HTTP endpoints, cloud load balancers can be placed in front of multiple seeder instances, distributing traffic and improving throughput.
      * **Cloud Object Storage Integration**: For seeders, consider storing files in cloud object storage (e.g., AWS S3, Google Cloud Storage, DigitalOcean Spaces) rather than local disk. This makes seeder instances stateless and highly scalable, as they can retrieve chunks from a shared, highly available storage layer.
      * **NAT Traversal (for true P2P)**: For a fully decentralized model where users can seed from home, implement NAT traversal techniques (e.g., UPnP, PCP, STUN/TURN servers) to enable direct peer-to-peer connections despite network address translation. This is a complex but crucial step for true P2P.
      * **Redundancy and High Availability**: Design for multiple instances of seeders and trackers to ensure continuous service availability and fault tolerance.

By implementing these future requirements, GoTorrent aims to evolve from a local demonstration into a robust, publicly accessible, and scalable file-sharing platform.

```
```
