# Eko ✦

**Eko** is an AI-powered snapshot versioning CLI designed to help you capture, restore, and visualize the evolution of your projects. Think of it as a "Time Machine" for your local development environment.

## Features

- **Snapshots:** Instantly save the state of your project.
- **Restore:** Revert your project to any previous snapshot with a single command.
- **Visual History:** A beautiful web-based timeline to explore your project's past.
- **Database Driven:** Efficiently tracks snapshots and metadata.

## Getting Started

### Prerequisites

- Go 1.26+
- Node.js (for the UI)

### Installation

```bash
go build -o eko main.go
```

### Usage

1. **Initialize Eko in your project:**
   ```bash
   eko init
   ```

2. **Save a snapshot:**
   ```bash
   eko save
   ```

3. **View history:**
   ```bash
   eko history
   ```

4. **Restore a snapshot:**
   ```bash
   eko restore <snapshot-id>
   ```

5. **Open the UI:**
   ```bash
   eko ui
   ```
   *(Note: You will need to run `npm run dev` in the `ui` directory for the web app to work)*

## Development

### UI
The frontend is built with Next.js and is located in the `ui/` directory.

```bash
cd ui
npm install
npm run dev
```

### API
The backend is a Go application that serves a REST API on port 7700.

## License
MIT
