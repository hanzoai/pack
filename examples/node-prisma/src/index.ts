import { PrismaClient, User, Post } from "@prisma/client";

// Ensure types are available
const user = null as any as User;
const post = null as any as Post;

const prisma = new PrismaClient();
console.log("Prisma client created", prisma);
