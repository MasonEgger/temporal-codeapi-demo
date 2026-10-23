// ABOUTME: Shared behavior for the playground tutorial pages: renders the tutorial Markdown, then wires up run consoles,
// the network switch, Web UI panels, and quizzes. Code snippets come live from examples/ via @@@SNIPSTART markers.
(async () => {
    const body = document.body;
    const language = body.dataset.language;
    const sandbox = body.dataset.sandbox;
    const codapiURL = body.dataset.codapiUrl || `${location.protocol}//${location.hostname}:1313/v1`;

    // One random session per browser tab ties runs, the network switch, and the Web UI together.
    function sessionID() {
        try {
            const saved = sessionStorage.getItem("temporal-playground-session");
            if (saved) return saved;
            const created = crypto.randomUUID();
            sessionStorage.setItem("temporal-playground-session", created);
            return created;
        } catch {
            return crypto.randomUUID();
        }
    }
    const session = sessionID();
    const sessionAPI = `/api/sessions/${session}`;

    function element(tag, className, text) {
        const node = document.createElement(tag);
        if (className) node.className = className;
        if (text !== undefined) node.textContent = text;
        return node;
    }

    // ---- Example files and snippets -----------------------------------------------

    const prismLanguages = { py: "python", go: "go", java: "java", ts: "typescript", xml: "markup" };
    const snippetStart = /@@@SNIPSTART\s+(\S+)/;

    // Load every example file the playground sends with a run.
    const files = new Map();
    await Promise.all(
        body.dataset.files.trim().split(/\s+/).map(async (path) => {
            const response = await fetch(`/examples/${language}/${path}`);
            if (!response.ok) throw new Error(`Could not load ${path} (${response.status})`);
            files.set(path, await response.text());
        }),
    );

    // Index the regions between @@@SNIPSTART <id> and @@@SNIPEND comments, like Temporal's snipsync tool.
    const snippets = new Map();
    for (const [path, text] of files) {
        let current = null;
        for (const line of text.split("\n")) {
            const start = line.match(snippetStart);
            if (start) {
                current = { path, lines: [] };
                snippets.set(start[1], current);
            } else if (line.includes("@@@SNIPEND")) {
                current = null;
            } else if (current) {
                current.lines.push(line);
            }
        }
    }

    function renderSnippet(id) {
        const snippet = snippets.get(id);
        if (!snippet) return element("p", "error", `Missing code snippet: ${id}`);
        const text = snippet.lines.join("\n").replace(/^\n+|\s+$/g, "") + "\n";
        const wrapper = element("div", "tutorial-file");
        const header = element("div", "file-header");
        header.append(element("span", "file-name", snippet.path));
        const copy = element("button", "file-copy", "Copy");
        copy.type = "button";
        copy.addEventListener("click", async () => {
            await navigator.clipboard.writeText(text);
            copy.textContent = "Copied";
            setTimeout(() => (copy.textContent = "Copy"), 1500);
        });
        header.append(copy);
        const pre = element("pre", "code-block");
        const code = element("code", `language-${prismLanguages[snippet.path.split(".").pop()] || "none"}`, text);
        pre.append(code);
        wrapper.append(header, pre);
        return wrapper;
    }

    // ---- Tutorial Markdown ------------------------------------------------------------

    // Blocks that marked can't express are swapped for placeholders, rendered separately, and put back.
    const placeholders = [];
    function placeholder(node) {
        placeholders.push(node);
        return `\n\n%%PLAYGROUND-${placeholders.length - 1}%%\n\n`;
    }

    function preprocess(markdown) {
        const imagesDir = body.dataset.images || language;
        // Docusaurus builds the page's H1 from the front matter title.
        const title = markdown.match(/^---\n[\s\S]*?^title:\s*(.+)$[\s\S]*?\n---\n/m);
        if (title) document.title = `Temporal Playground: ${title[1].trim()}`;
        return markdown
            // Docusaurus front matter (replaced by its title) and the SDK banner image.
            .replace(/^---\n[\s\S]*?\n---\n/, title ? `# ${title[1].trim()}\n\n` : "")
            .replace(/^<img className="banner".*$/m, "")
            // Code snippets come from the example files, not the copies pasted into the Markdown.
            .replace(/<!--SNIPSTART\s+(\S+?)\s*-->[\s\S]*?<!--SNIPEND-->/g, (_, id) => placeholder(renderSnippet(id)))
            // Admonitions (:::note Title ... :::) and the playground's :::setup accordion.
            .replace(/^:::(\w+)[ \t]*(.*)\n([\s\S]*?)\n:::[ \t]*$/gm, (_, kind, title, inner) => {
                if (kind === "setup") {
                    const details = element("details", "setup");
                    details.innerHTML = `<summary></summary>${marked.parse(inner)}`;
                    details.querySelector("summary").textContent = title;
                    return placeholder(details);
                }
                const note = element("div", "note");
                note.innerHTML = (title ? "<p class=\"note-title\"></p>" : "") + marked.parse(inner);
                if (title) note.querySelector(".note-title").textContent = title;
                return placeholder(note);
            })
            // Screenshots live next to the pages instead of next to the original tutorial.
            .replace(/\]\(images\//g, `](tutorials/images/${imagesDir}/`);
    }

    const main = document.querySelector("main");
    try {
        const response = await fetch(body.dataset.tutorial);
        if (!response.ok) throw new Error(`Could not load the tutorial (${response.status})`);
        main.innerHTML = marked.parse(preprocess(await response.text()));
    } catch (error) {
        main.replaceChildren(element("p", "error", error.message));
        return;
    }
    // Placeholders can nest (a snippet inside a :::setup block), so repeat until none are left.
    for (let replaced = true; replaced; ) {
        replaced = false;
        for (const paragraph of main.querySelectorAll("p")) {
            const match = paragraph.textContent.trim().match(/^%%PLAYGROUND-(\d+)%%$/);
            if (match) {
                paragraph.replaceWith(placeholders[Number(match[1])]);
                replaced = true;
            }
        }
    }

    // Fenced blocks: ```command becomes a shell block and ```output a sample-output block.
    for (const code of main.querySelectorAll("pre > code")) {
        const pre = code.parentElement;
        if (code.classList.contains("language-command")) {
            pre.className = "shell";
            code.className = "language-bash";
        } else if (code.classList.contains("language-output")) {
            pre.className = "sample";
        } else if (!pre.classList.contains("code-block")) {
            pre.className = "shell";
        }
    }
    if (window.Prism) Prism.highlightAllUnder(main);
    for (const heading of main.querySelectorAll("h2, h3")) {
        heading.id = heading.textContent.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "");
    }

    // ---- Learner name ------------------------------------------------------------

    let learnerName = "Ziggy";
    function syncNameInputs(source) {
        for (const input of document.querySelectorAll(".run-name input")) {
            if (input !== source) input.value = learnerName;
        }
    }

    // ---- Running programs ----------------------------------------------------------

    async function execute(command, variant) {
        const request = {};
        for (const [path, text] of files) {
            request[path] = text;
        }
        if (variant) {
            const original = request[variant.file];
            if (!original || !original.includes(variant.find)) {
                throw new Error(`The exercise could not find the line to change in ${variant.file}.`);
            }
            request[variant.file] = original.replace(variant.find, variant.replace);
        }
        request["session.id"] = session;
        request["name.txt"] = learnerName;

        // No client-side timeout: a run waiting out a network outage can take a couple of minutes,
        // and the server enforces its own limit on every run.
        const response = await fetch(`${codapiURL}/exec`, {
            method: "POST",
            headers: { accept: "application/json", "content-type": "application/json" },
            body: JSON.stringify({ sandbox, command, files: request }),
        });
        if (response.status === 429) {
            throw new Error("The playground is busy running other programs. Try again in a few seconds.");
        }
        const result = await response.json().catch(() => null);
        if (!response.ok) {
            throw new Error(result?.stderr || `The playground could not run this program (HTTP ${response.status}).`);
        }
        return result;
    }

    // When a program exits with an error, Codapi moves all of its output into stderr and appends
    // "(exit status N)". Pull that suffix off so the console header can show it.
    function splitExitStatus(text) {
        const match = text.match(/\s*\(exit status (\d+)\)\s*$/);
        return match ? { text: text.slice(0, match.index), exitStatus: match[1] } : { text, exitStatus: null };
    }

    // Split the harness output into its labeled sections: program, Worker, and each Event History.
    function outputSections(text, label) {
        const sections = [];
        let current = { title: label, lines: [] };
        for (const line of text.split("\n")) {
            const marker = line.match(/^=== (.+) ===$/);
            if (marker) {
                sections.push(current);
                current = { title: marker[1], lines: [] };
            } else {
                current.lines.push(line);
            }
        }
        sections.push(current);
        return sections.filter((section) => section.lines.join("").trim());
    }

    // Lines that report a failure get the failure color; everything else stays neutral.
    const failurePattern = /(FAILED|FAIL\b|Error\b|Exception\b|Traceback|failing|✖|✘|panic:)/;
    function renderLines(pre, lines) {
        for (const line of lines.join("\n").replace(/\n+$/, "").split("\n")) {
            const span = element("span", failurePattern.test(line) ? "line fail" : "line", line + "\n");
            pre.append(span);
        }
    }

    function renderConsole(consoleNode, state, detail) {
        consoleNode.hidden = false;
        consoleNode.dataset.state = state;
        const status = consoleNode.querySelector(".console-status");
        status.textContent = detail;
    }

    for (const block of document.querySelectorAll(".run-block")) {
        const command = block.dataset.command;
        const outputLabel = block.dataset.outputLabel || "Output";
        const variant = block.dataset.variantFile
            ? { file: block.dataset.variantFile, find: block.dataset.variantFind, replace: block.dataset.variantReplace }
            : null;

        const controls = element("div", "run-controls");
        const buttons = [];
        function addButton(label, buttonCommand, primary) {
            const button = element("button", primary ? "run-button primary" : "run-button", label);
            button.type = "button";
            button.addEventListener("click", () => run(buttonCommand));
            buttons.push(button);
            controls.append(button);
        }
        addButton(block.dataset.label || "Run", command, true);
        if (block.dataset.history !== undefined) addButton("Run with Event History", "history", false);

        if (block.dataset.name !== undefined) {
            const nameField = element("label", "run-name");
            nameField.append(element("span", null, "Name"));
            const input = element("input");
            input.type = "text";
            input.maxLength = 60;
            input.value = learnerName;
            input.addEventListener("input", () => {
                learnerName = input.value.trim() || "Ziggy";
                syncNameInputs(input);
            });
            nameField.append(input);
            controls.append(nameField);
        }

        // Warn near every Run button while the network is off, so a run that keeps retrying isn't a mystery.
        const networkNote = element("p", "network-off-note");
        networkNote.append(element("span", null, "The network is off, so this run will keep retrying the Activities until you turn it back on."));
        const turnOn = element("button", "network-on-button", "Turn the network on");
        turnOn.type = "button";
        turnOn.addEventListener("click", () => setNetwork(true));
        networkNote.append(turnOn);

        const consoleNode = element("div", "console");
        consoleNode.hidden = true;
        consoleNode.setAttribute("aria-live", "polite");
        const consoleHeader = element("div", "console-header");
        consoleHeader.append(element("span", "console-dot"), element("span", "console-status"));
        const consoleBody = element("div", "console-body");
        consoleNode.append(consoleHeader, consoleBody);
        block.append(controls, networkNote, consoleNode);

        async function run(runCommand) {
            for (const button of buttons) button.disabled = true;
            consoleBody.replaceChildren();
            const started = Date.now();
            const tick = () => renderConsole(consoleNode, "running", `Running… ${Math.round((Date.now() - started) / 1000)}s`);
            tick();
            const timer = setInterval(tick, 1000);

            // While the program runs, report failed Activity attempts from the run's own Temporal Service.
            const retries = runCommand === "test" ? null : watchRetries(consoleBody, started);
            try {
                const result = await execute(runCommand, variant);
                retries?.stop(result.ok);
                const seconds = (result.duration / 1000).toFixed(1);
                const stdout = result.stdout || "";
                const { text: stderr, exitStatus } = splitExitStatus(result.stderr || "");
                const exitNote = exitStatus ? ` (exit code ${exitStatus})` : "";
                renderConsole(consoleNode, result.ok ? "done" : "failed", result.ok ? `Finished in ${seconds}s` : `Failed after ${seconds}s${exitNote}`);

                // A successful run keeps its output in stdout; a failed one has it all in stderr.
                const sections = outputSections(stdout.trim() ? stdout : stderr, outputLabel);
                if (stdout.trim() && stderr.trim()) sections.push({ title: "Errors", lines: stderr.split("\n") });
                for (const section of sections) {
                    // The Worker's raw log (stack traces for every failed attempt) stays available but folded away.
                    const isWorkerLog = section.title === "Worker output";
                    const sectionNode = element(isWorkerLog ? "details" : "section", "console-section");
                    sectionNode.append(element(isWorkerLog ? "summary" : "h4", null, isWorkerLog ? "Worker log" : section.title));
                    const pre = element("pre");
                    renderLines(pre, section.lines);
                    sectionNode.append(pre);
                    consoleBody.append(sectionNode);
                }
            } catch (error) {
                retries?.stop(false);
                renderConsole(consoleNode, "failed", "Could not run");
                consoleBody.append(element("p", "console-message", error.message));
            } finally {
                clearInterval(timer);
                for (const button of buttons) button.disabled = false;
                refreshStatus();
            }
        }
    }

    // Poll the run's Temporal Service (through the Web UI's API) and print one line per failed Activity attempt.
    function watchRetries(consoleBody, started) {
        const api = `/ui/${session}/api/v1/namespaces/default`;
        const section = element("section", "console-section retries");
        section.append(element("h4", null, "Temporal"));
        const list = element("pre");
        section.hidden = true;
        section.append(list);
        consoleBody.prepend(section);
        const reported = new Set();
        let stopped = false;

        async function poll() {
            try {
                const running = await (await fetch(`${api}/workflows?query=${encodeURIComponent('ExecutionStatus="Running"')}`)).json();
                const execution = running.executions?.[0]?.execution;
                if (!execution) return;
                const startTime = Date.parse(running.executions[0].startTime || 0);
                if (startTime && startTime < started - 5000) return; // an older run's Workflow, not this one
                const described = await (await fetch(`${api}/workflows/${encodeURIComponent(execution.workflowId)}`)).json();
                for (const pending of described.pendingActivities || []) {
                    if (!pending.lastFailure) continue;
                    // A scheduled retry's attempt number is the next try, so the failed one is attempt - 1.
                    const scheduled = pending.state === "PENDING_ACTIVITY_STATE_SCHEDULED";
                    const failedAttempt = scheduled ? pending.attempt - 1 : pending.attempt;
                    const key = `${pending.activityType?.name}:${failedAttempt}`;
                    if (failedAttempt < 1 || reported.has(key)) continue;
                    reported.add(key);
                    // Every SDK reports the playground's blocked request as a 503 (Node says "fetch failed").
                    const message = pending.lastFailure.message || "";
                    const blocked = !status.networkOn || /503|Service Unavailable|fetch failed/i.test(message);
                    const reason = blocked ? "the network is unavailable" : message.slice(0, 160);
                    const wait = scheduled && pending.nextAttemptScheduleTime
                        ? Math.max(1, Math.round((Date.parse(pending.nextAttemptScheduleTime) - Date.now()) / 1000))
                        : null;
                    const retry = wait ? ` Temporal will retry in ${wait}s.` : " Temporal will retry it.";
                    list.append(element("span", "line", `${pending.activityType?.name} failed on attempt ${failedAttempt}: ${reason}.${retry}\n`));
                    section.hidden = false;
                }
            } catch {
                // The run's Temporal Service may not be up yet; the next poll tries again.
            }
        }

        const interval = setInterval(() => { if (!stopped) poll(); }, 1000);
        return {
            stop(succeeded) {
                stopped = true;
                clearInterval(interval);
                if (reported.size && succeeded) {
                    list.append(element("span", "line ok", "The next attempt succeeded, and the Workflow completed.\n"));
                }
            },
        };
    }

    // ---- Network switch ------------------------------------------------------------

    const switches = [...document.querySelectorAll(".network-switch")];
    let status = { networkOn: true, blocked: 0, hasRun: false };

    function renderStatus() {
        for (const container of switches) {
            const button = container.querySelector("button");
            button.setAttribute("aria-checked", String(status.networkOn));
            container.dataset.state = status.networkOn ? "on" : "off";
            container.querySelector(".switch-state").textContent = status.networkOn ? "On" : "Off";
            container.querySelector(".switch-detail").textContent = status.networkOn
                ? ""
                : `${status.blocked} request${status.blocked === 1 ? "" : "s"} blocked`;
        }
        for (const panel of document.querySelectorAll(".webui")) {
            panel.dataset.hasRun = String(status.hasRun);
        }
        for (const block of document.querySelectorAll(".run-block")) {
            block.dataset.network = status.networkOn ? "on" : "off";
        }
    }

    async function setNetwork(on) {
        const response = await fetch(`${sessionAPI}/network`, {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ on }),
        });
        if (response.ok) {
            status = await response.json();
            renderStatus();
        }
    }

    async function refreshStatus() {
        try {
            const response = await fetch(sessionAPI);
            if (response.ok) {
                status = await response.json();
                renderStatus();
            }
        } catch {
            // The gateway may be restarting; the next poll catches up.
        }
    }

    for (const container of switches) {
        container.innerHTML = `
            <span class="switch-label">Network</span>
            <button type="button" role="switch" aria-checked="true" aria-label="Network access for your runs">
                <span class="switch-track"><span class="switch-thumb"></span></span>
                <span class="switch-state">On</span>
            </button>
            <span class="switch-detail" aria-live="polite"></span>`;
        container.querySelector("button").addEventListener("click", () => setNetwork(!status.networkOn));
    }
    renderStatus();
    refreshStatus();
    setInterval(refreshStatus, 1500);

    // ---- Web UI panels -----------------------------------------------------------------

    for (const panel of document.querySelectorAll(".webui")) {
        const uiPath = `/ui/${session}/namespaces/default/workflows`;
        const bar = element("div", "webui-bar");
        bar.append(element("strong", null, "Temporal Web UI"), element("span", "webui-hint", "Your most recent run's Temporal Service"));
        const load = element("button", "run-button primary", "Open the Web UI");
        load.type = "button";
        const newTab = element("a", "webui-tab", "Open in a new tab");
        newTab.href = uiPath;
        newTab.target = "_blank";
        newTab.rel = "noopener";
        bar.append(load, newTab);
        const empty = element("p", "webui-empty", "Run the program first. Each run starts its own Temporal Service, and this panel shows it.");
        panel.append(bar, empty);

        load.addEventListener("click", () => {
            // The Web UI shares this page's origin, so collapse its sidebar on first visit to leave
            // more room in the panel. A learner who expands it again keeps their choice.
            try {
                if (localStorage.getItem("navOpen") === null) localStorage.setItem("navOpen", "false");
            } catch {
                // Storage can be blocked; the sidebar then opens expanded, which still works.
            }
            let frame = panel.querySelector("iframe");
            if (!frame) {
                frame = element("iframe");
                frame.title = "Temporal Web UI";
                panel.append(frame);
            }
            frame.src = uiPath;
            load.textContent = "Reload the Web UI";
        });
    }

    // ---- Quizzes ---------------------------------------------------------------------

    document.querySelectorAll(".quiz").forEach((quiz, quizIndex) => {
        const options = [...quiz.querySelectorAll("li")];
        const list = quiz.querySelector("ol");
        const choices = element("div", "quiz-choices");
        choices.setAttribute("role", "radiogroup");
        options.forEach((option, optionIndex) => {
            const label = element("label", "quiz-choice");
            const input = element("input");
            input.type = "radio";
            input.name = `quiz-${quizIndex}`;
            input.value = String(optionIndex);
            input.dataset.correct = String(option.dataset.correct !== undefined);
            label.append(input, element("span", null, option.textContent.trim()));
            choices.append(label);
        });
        list.replaceWith(choices);

        const check = element("button", "run-button primary", "Check answer");
        check.type = "button";
        const feedback = quiz.querySelector(".quiz-explanation");
        feedback.hidden = true;
        const verdict = element("p", "quiz-verdict");
        verdict.hidden = true;
        quiz.append(check, verdict);
        feedback.remove();
        quiz.append(feedback);

        check.addEventListener("click", () => {
            const picked = choices.querySelector("input:checked");
            if (!picked) {
                verdict.hidden = false;
                verdict.dataset.result = "none";
                verdict.textContent = "Pick an answer first.";
                return;
            }
            const correct = picked.dataset.correct === "true";
            for (const label of choices.querySelectorAll(".quiz-choice")) {
                const input = label.querySelector("input");
                label.dataset.result = input.dataset.correct === "true" ? "correct" : input === picked ? "wrong" : "";
            }
            verdict.hidden = false;
            verdict.dataset.result = correct ? "correct" : "wrong";
            verdict.textContent = correct ? "Correct." : "Not quite. The correct answer is highlighted.";
            feedback.hidden = false;
        });
    });

    // ---- Language switcher -----------------------------------------------------------

    for (const link of document.querySelectorAll(".language-switcher a")) {
        if (link.dataset.language === language) link.setAttribute("aria-current", "page");
    }
})();
