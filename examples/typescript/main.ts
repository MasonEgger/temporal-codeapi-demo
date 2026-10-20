// ABOUTME: Hello World for the temporal-typescript Codapi sandbox.
// Runs a Worker and Client against the per-run dev server and prints the Workflow result.
import { randomUUID } from 'node:crypto';
import { fileURLToPath } from 'node:url';
import { Client, Connection } from '@temporalio/client';
import { NativeConnection, Worker } from '@temporalio/worker';
import * as activities from './activities.ts';
import { helloWorkflow } from './workflows.ts';

const TASK_QUEUE = 'hello-task-queue';

const worker = await Worker.create({
  connection: await NativeConnection.connect({ address: 'localhost:7233' }),
  taskQueue: TASK_QUEUE,
  workflowsPath: fileURLToPath(new URL('./workflows.ts', import.meta.url)),
  activities,
});
const client = new Client({ connection: await Connection.connect({ address: 'localhost:7233' }) });

// Run the Worker only until the Workflow finishes, then shut it down.
const result = await worker.runUntil(
  client.workflow.execute(helloWorkflow, {
    args: ['Temporal'],
    workflowId: `hello-workflow-${randomUUID()}`,
    taskQueue: TASK_QUEUE,
  }),
);
console.log(`Workflow result: ${result}`);
