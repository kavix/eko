export type Snapshot = {
  id: string;
  timestamp: string;
  prompt: string;
  aiSummary: string;
  filesChanged: string[];
  model: string;
};

export type Project = {
  id: string;
  name: string;
  snapshots: Snapshot[];
};

export const PROJECTS: Project[] = [
  {
    id: "my-app",
    name: "my-app",
    snapshots: [
      {
        id: "s1",
        timestamp: "2026-06-15 01:05",
        prompt: "Add JWT authentication middleware",
        aiSummary: "Created middleware/auth.ts, added token validation, updated route guards across all protected endpoints.",
        filesChanged: ["middleware/auth.ts", "routes/user.ts", "routes/admin.ts", "lib/jwt.ts", "types/auth.ts"],
        model: "Claude Sonnet 4.6",
      },
      {
        id: "s2",
        timestamp: "2026-06-15 00:47",
        prompt: "Create user schema with Prisma",
        aiSummary: "Defined User model with email, passwordHash, role, and timestamps. Added migration file.",
        filesChanged: ["prisma/schema.prisma", "prisma/migrations/001_user.sql"],
        model: "Claude Sonnet 4.6",
      },
      {
        id: "s3",
        timestamp: "2026-06-15 00:32",
        prompt: "Scaffold REST API routes for /auth",
        aiSummary: "Added login, register, and logout endpoints. Wired to Prisma client and bcrypt for password hashing.",
        filesChanged: ["routes/auth.ts", "controllers/authController.ts", "lib/bcrypt.ts"],
        model: "Claude Sonnet 4.6",
      },
      {
        id: "s4",
        timestamp: "2026-06-14 23:58",
        prompt: "Set up Express server with TypeScript",
        aiSummary: "Initialized Express with TypeScript config, added ts-node-dev for hot reload, configured path aliases.",
        filesChanged: ["server.ts", "tsconfig.json", "package.json", ".env.example"],
        model: "Claude Sonnet 4.6",
      },
      {
        id: "s5",
        timestamp: "2026-06-14 23:41",
        prompt: "Initialize project with basic folder structure",
        aiSummary: "Created standard MVC folder layout: routes/, controllers/, middleware/, lib/, types/.",
        filesChanged: ["README.md", ".gitignore", "package.json"],
        model: "Claude Sonnet 4.6",
      },
    ],
  },
  {
    id: "ecommerce-ai",
    name: "ecommerce-ai",
    snapshots: [
      {
        id: "e1",
        timestamp: "2026-06-15 00:55",
        prompt: "Add product recommendation engine",
        aiSummary: "Implemented collaborative filtering using cosine similarity. Stores user interaction vectors in Redis.",
        filesChanged: ["lib/recommend.ts", "api/recommend.ts", "workers/vectorize.ts"],
        model: "Claude Sonnet 4.6",
      },
      {
        id: "e2",
        timestamp: "2026-06-15 00:30",
        prompt: "Build shopping cart with localStorage persistence",
        aiSummary: "Cart state managed with Zustand, persisted to localStorage via middleware. Syncs on mount.",
        filesChanged: ["stores/cart.ts", "components/Cart.tsx", "components/CartItem.tsx"],
        model: "Claude Sonnet 4.6",
      },
      {
        id: "e3",
        timestamp: "2026-06-15 00:10",
        prompt: "Integrate Stripe checkout",
        aiSummary: "Added stripe/stripe-js, created checkout session API route, redirect to Stripe hosted page.",
        filesChanged: ["api/checkout.ts", "lib/stripe.ts", "pages/success.tsx"],
        model: "Claude Sonnet 4.6",
      },
      {
        id: "e4",
        timestamp: "2026-06-14 23:45",
        prompt: "Create product listing page with filters",
        aiSummary: "Server-side filtered product list with category, price range, and sort. Paginated with cursor.",
        filesChanged: ["pages/products.tsx", "components/FilterPanel.tsx", "api/products.ts"],
        model: "Claude Sonnet 4.6",
      },
      {
        id: "e5",
        timestamp: "2026-06-14 23:20",
        prompt: "Set up Next.js with Tailwind and Prisma",
        aiSummary: "Bootstrapped project, configured Tailwind, initialized Prisma with PostgreSQL adapter.",
        filesChanged: ["tailwind.config.ts", "prisma/schema.prisma", "next.config.ts"],
        model: "Claude Sonnet 4.6",
      },
    ],
  },
  {
    id: "chat-app",
    name: "chat-app",
    snapshots: [
      {
        id: "c1",
        timestamp: "2026-06-15 01:02",
        prompt: "Add real-time typing indicators",
        aiSummary: "Implemented presence events via Socket.IO. Debounced emit on keypress, auto-clear after 2s.",
        filesChanged: ["socket/presence.ts", "components/TypingIndicator.tsx"],
        model: "Claude Sonnet 4.6",
      },
      {
        id: "c2",
        timestamp: "2026-06-15 00:40",
        prompt: "Implement message threading",
        aiSummary: "Added parentId to Message schema, recursive thread rendering, collapsed/expanded state.",
        filesChanged: ["models/message.ts", "components/Thread.tsx", "api/messages.ts"],
        model: "Claude Sonnet 4.6",
      },
      {
        id: "c3",
        timestamp: "2026-06-15 00:18",
        prompt: "Build real-time chat with Socket.IO",
        aiSummary: "WebSocket server with room-based message routing. Client hook useSocket wraps connection lifecycle.",
        filesChanged: ["server/socket.ts", "hooks/useSocket.ts", "components/ChatWindow.tsx"],
        model: "Claude Sonnet 4.6",
      },
      {
        id: "c4",
        timestamp: "2026-06-14 23:55",
        prompt: "Create message input with emoji picker",
        aiSummary: "Native emoji picker via browser EmojiPicker API. Falls back to unicode insert on unsupported browsers.",
        filesChanged: ["components/MessageInput.tsx", "lib/emoji.ts"],
        model: "Claude Sonnet 4.6",
      },
      {
        id: "c5",
        timestamp: "2026-06-14 23:30",
        prompt: "Initialize chat app with Next.js",
        aiSummary: "Scaffold with App Router. Set up Socket.IO server in custom Next.js server.ts.",
        filesChanged: ["server.ts", "app/layout.tsx", "next.config.ts"],
        model: "Claude Sonnet 4.6",
      },
    ],
  },
];
