<script lang="ts">
    import "fluent-svelte/theme.css";
    import {
        ProcessData,
        ReadSettings,
        SelectFiles,
        SelectOutputFolder,
        SelectWatermark
    } from '../wailsjs/go/main/App.js'
    import {BrowserOpenURL, EventsOnMultiple, LogInfo} from '../wailsjs/runtime/runtime'
    import {
        Button,
        ContentDialog,
        ListItem,
        NumberBox,
        ProgressBar,
        RadioButton,
        TextBlock,
        TextBox
    } from "fluent-svelte";
    import Delete from "@fluentui/svg-icons/icons/delete_24_regular.svg?raw";


    interface ConvertFile {
        name: string;
        path: string;
        type: string;
    }

    interface ProgressData {
        currentFile: ConvertFile;
        percentage: number;
        running: boolean;
        error: string;
    }

    interface ProcessData {
        files: Array<ConvertFile>;
        transparent: number;
        size: number;
        watermark: string;
        prefix: string;
        position: string;
    }

    let prefix: string = "wm_";
    let watermark: string;
    let transparent: number = 0;
    let files = [];
    let size = 100;
    let outputFolder = "";
    let position: string = "top-left";

    let status = {} as ProgressData;

    (function readSettings() {
        ReadSettings().then(result => {
            if (result) {
                LogInfo("Loaded settings file")
                const data = JSON.parse(result);
                const parsedData = data as ProcessData;
                prefix = parsedData.prefix;
                watermark = parsedData.watermark;
                transparent = parsedData.transparent;
                size = parsedData.size;
                position = parsedData.position;
            }
        });
    }());

    function selectFiles(): void {
        SelectFiles();
    }

    function clearFiles(): void {
        files = [];
    }

    function selectWatermark(): void {
        SelectWatermark();
    }

    function selectOutputFolder(): void {
        SelectOutputFolder();
    }

    function startProcessing(): void {
        status = {} as ProgressData;
        ProcessData({
            watermark,
            transparent,
            prefix,
            files,
            size,
            position,
            outputFolder
        });
    }

    EventsOnMultiple("selectWatermark", data => {
        watermark = data;
    }, undefined);

    EventsOnMultiple("selectOutputFolder", data => {
        outputFolder = data;
    }, undefined);

    EventsOnMultiple("selectFiles", data => {
        files = JSON.parse(data)
    }, undefined);

    EventsOnMultiple("progress", data => {
        data = JSON.parse(data);
        let parsedData = data as ProgressData;
        LogInfo(JSON.stringify(parsedData));
        status = parsedData;
    }, undefined);

</script>

<main>
    <div class="grid-files">
        <TextBlock variant="subtitle">Dateien</TextBlock>

        <div class="files">
            {#if files.length > 0}
                {#each files as f}
                    <ListItem>
                        <svg slot="icon" width="16" height="16" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
                            <path
                                d={f.type === "img" ? "M4.5 2A2.5 2.5 0 002 4.5v7c0 .51.15.98.41 1.38L6.8 8.49a1.7 1.7 0 012.4 0l4.39 4.39c.26-.4.41-.87.41-1.38v-7A2.5 2.5 0 0011.5 2h-7zm7 3.5a1 1 0 11-2 0 1 1 0 012 0zm1.38 8.09L8.5 9.2a.7.7 0 00-1 0L3.13 13.6c.4.26.87.41 1.38.41h7c.5 0 .98-.15 1.38-.41z" : "M13.22 4.25L7.09 6.24a.5.5 0 01-.24.08L4.75 7h8.75c.28 0 .5.22.5.5v5a2.5 2.5 0 01-2.5 2.5h-7A2.5 2.5 0 012 12.5v-5c0-.1.03-.2.09-.29l-.2-.6A2.5 2.5 0 013.5 3.46l6.66-2.16a2.5 2.5 0 013.15 1.6l.23.72a.5.5 0 01-.32.63zm-3.75.16l1.29-2.23-.3.07-1.24.4-1.3 2.27 1.55-.5zm2.3-1.98l-.02.04-.85 1.48 1.53-.5-.07-.24a1.5 1.5 0 00-.6-.78zm-3.97.69l-1.56.5-1.3 2.27 1.55-.51 1.3-2.26zM2.92 6.54l.59-.2 1.3-2.26-1 .33a1.5 1.5 0 00-.96 1.9l.07.23z"}
                                fill="currentColor"
                            />
                        </svg>

                        {f.name}
                    </ListItem>
                {/each}
            {:else}
                <div class="files-empty" on:click={selectFiles}>
                    <div>
                        Keine Dateien
                    </div>
                </div>
            {/if}
        </div>
        <Button variant="hyperlink" class="clear" disabled={files.length === 0 || status.running}
                on:click={clearFiles}>{@html Delete}
            <TextBlock>Alle löschen</TextBlock>
        </Button>
    </div>
    <div class="grid-settings">
        <TextBlock variant="subtitle">Einstellungen</TextBlock>
        <br/>
        <TextBlock variant="bodyLarge">Wassermarke</TextBlock>
        <div class="watermark">
            <TextBox bind:value={watermark} placeholder="Auswählen..."/>
            <Button on:click={selectWatermark}>...</Button>
        </div>
        <br/>
        <div class="grid-numbers">
            <TextBlock variant="bodyLarge" class="transparent-title">Transparenz</TextBlock>
            <NumberBox
                placeholder="%"
                clearButton={false}
                inline
                min={0}
                max={100}
                bind:value={transparent}
                class="transparent-value"
            />
            <TextBlock variant="bodyLarge" class="size-title">Größe</TextBlock>
            <NumberBox
                placeholder="%"
                clearButton={false}
                inline
                min={0}
                max={100}
                bind:value={size}
                class="size-value"
            />
        </div>
        <br/>
        <TextBlock variant="bodyLarge">Position</TextBlock>
        <div class="position">
            <RadioButton bind:group={position} value="top-left">Oben Links</RadioButton>
            <RadioButton bind:group={position} value="top-right">Oben Rechts</RadioButton>
            <RadioButton bind:group={position} value="bottom-left">Unten Links</RadioButton>
            <RadioButton bind:group={position} value="bottom-right">Unten Rechts</RadioButton>
        </div>
        <br/>
        <TextBlock variant="bodyLarge">Präfix</TextBlock>
        <TextBox bind:value={prefix} placeholder="Export Präfix"/>
        <br/>
        <TextBlock variant="bodyLarge">Ausgabe Ordner</TextBlock>
        <div class="watermark">
            <TextBox bind:value={outputFolder} placeholder="Leer für gleicher Ordner"/>
            <Button on:click={selectOutputFolder}>...</Button>
        </div>
    </div>
    <div class="grid-button">
        {#if status.running === true}
            <TextBlock variant="caption">Verarbeite {status.currentFile.name}...</TextBlock>
            <ProgressBar value={status.percentage}/>
        {/if}
        <Button variant="accent" disabled={files.length === 0 || !watermark || status.running}
                on:click={startProcessing}>Wasserzeichen hinzufügen
        </Button>
    </div>
    {#if status.error}
        <ContentDialog open title="Fehler!" on:backdropclick={() => status.error = ""}>{status.error}</ContentDialog>
    {/if}
</main>
<footer on:click={e => {
        e.preventDefault();
        BrowserOpenURL("https://mc8051.de/?ref=wm-generator");
    }}>
    <TextBlock variant="caption">mc8051 coded with 💗 by Niklas Schütrumpf</TextBlock>
</footer>
