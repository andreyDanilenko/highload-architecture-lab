/**
 * Load env before any config: shared 01-atomic-inventory/.env first, then node/.env overrides.
 * So one .env at project root works for both Docker Compose and Node (and later Go).
 */
import dotenv from "dotenv";
import path from "path";

const cwd = process.cwd();
dotenv.config({ path: path.resolve(cwd, "../.env") });
dotenv.config();
