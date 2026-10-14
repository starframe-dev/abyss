import { fileURLToPath } from 'node:url'
import path from 'node:path'
import { runProblemCheck } from '@human-horizon/code-check/problem-check'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
// Pipeline is at .lore/weft/pipelines/code-check/queue-check.ts
// Project root is 4 levels up
const projectPath = path.resolve(__dirname, '../../../../')

export async function main(): Promise<void> {
    // Alias for check-problems: problem scanning is the queue item.
    const result = await runProblemCheck(projectPath)
    if (!result.ok) {
        console.error(result.error.message)
        process.exit(1)
    }
    console.log(JSON.stringify(result.value, null, 2))
}

await main()
