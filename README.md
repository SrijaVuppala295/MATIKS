# Matiks Leaderboard 🏆

A high-performance, real-time leaderboard application designed to handle millions of users with instant search lookup and live score updates.

![Matiks Leaderboard](frontend/assets/images/adaptive-icon.png)

## 🚀 Key Features

*   **Instant Search Lookup**: Find any user among millions in milliseconds using optimized Redis indexing (`ZSCAN/HSCAN`).
*   **Real-time Updates**: User scores and ranks update instantly across all connected clients.
*   **High Scalability**: Built to handle millions of concurrent users with ease, leveraging the speed of Redis and Golang.
*   **Premium UI**: A classy, responsive interface with glassmorphism effects, hover animations, and a dynamic podium presentation.
*   **Cross-Platform**: Fully responsive design that works seamlessly on Desktop and Mobile devices.

## 🛠 Tech Stack

### Backend
*   **Language**: [Golang](https://go.dev/) (High performance & concurrency)
*   **Database**: [Redis](https://redis.io/) (In-memory data structure store for caching & leaderboards)
*   **API**: RESTful API for frontend communication.

### Frontend
*   **Framework**: [React Native](https://reactnative.dev/) / [Expo](https://expo.dev/) (Web & Mobile)
*   **Styling**: Custom CSS/Styles with Glassmorphism aesthetic.
*   **Animations**: `Animated` API for smooth interactions.

## 📂 Project Structure

```
MATIKS/
├── backend/            # Golang Server & Business Logic
│   ├── main.go         # Entry point & API Handlers
│   ├── go.mod          # Go Dependencies
│   └── ...
├── frontend/           # React Native / Expo Client
│   ├── app/            # Application Routes & Screens
│   │   └── (tabs)/     # Main Tab Navigation
│   │       └── index.tsx # Leaderboard & Podium Screen
│   ├── assets/         # Images & Icons
│   ├── package.json    # JS Dependencies
│   └── ...
└── README.md           # Project Documentation
```

## ⚙️ Installation & Setup

### Prerequisites
*   [Go](https://go.dev/dl/) (v1.18+)
*   [Node.js](https://nodejs.org/) (v16+) & npm
*   [Redis](https://redis.io/download/) (Running locally or cloud)

### 1. Backend Setup
1.  Navigate to the backend directory:
    ```bash
    cd backend
    ```
2.  Install dependencies:
    ```bash
    go mod tidy
    ```
3.  Start the server (ensure Redis is running on default port `6379`):
    ```bash
    go run main.go
    ```
    *Server will start on `http://localhost:8080`*

### 2. Frontend Setup
1.  Navigate to the frontend directory:
    ```bash
    cd frontend
    ```
2.  Install dependencies:
    ```bash
    npm install
    ```
3.  Start the Expo development server:
    ```bash
    npm start
    ```
4.  Open the displayed URL (e.g., `http://localhost:8081`) in your browser or scan the QR code with your mobile device.

## 🧪 Testing

You can use **Postman** or `curl` to test the backend APIs directly.

### Endpoints

*   **Get Leaderboard**
    *   `GET /leaderboard?page=1&limit=50`
    *   Fetches the top 50 users.

*   **Search User**
    *   `GET /leaderboard?q=username`
    *   Instantly searches for a user by name.

*   **Update Score (Internal/Simulated)**
    *   The backend automatically simulates score updates for demonstration. You can verify this by observing the live rank changes on the frontend.

## 📱 Mobile Support
This project is fully responsive!
*   **On Mobile**: The header adapts to show a compact Logo icon, and the podium layouts adjust to fit small screens perfectly.
*   **On Desktop**: Enjoy the full "Premium" experience with hover effects and expansive layouts.

---
*Built with ❤️ for Matiks.*
