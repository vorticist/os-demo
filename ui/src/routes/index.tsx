import {$, component$} from "@builder.io/qwik";
import type {DocumentHead} from "@builder.io/qwik-city";
import {Form, routeAction$, routeLoader$, z, zod$} from "@builder.io/qwik-city";
import styles from "~/routes/demo/todolist/todolist.module.css";
import WebSocket from 'ws';

interface ListItem {
    text: string;
}

export const list: ListItem[] = [];

export const useListLoader = routeLoader$(() => {
    return list;
});


export default component$(() => {
    const list = useListLoader();
    const predict = usePredict();
    let url = ""
    return (
        <div class="container container-center">
                <input type="text" name="url" required class={styles.input} value={url}/>
                <button type="submit" class="button-dark" onClick$={predict}>
                    Predict
                </button>
            <p class={styles.hint}>
                Use a link to an image or a youtube video
            </p>
            <div class="container container-center">
                <details>
                    <summary>
                        <h3>Logs</h3>
                        <span class="icon">👇</span>
                    </summary>
                    {list.value.length === 0 ? (
                        <span class={styles.empty}>No items found</span>
                    ) : (
                        <ul class={styles.list}>
                            {list.value.map((item, index) => (
                                <li key={`items-${index}`}>{item.text}</li>
                            ))}
                        </ul>
                    )}
                </details>
            </div>
        </div>
    );
});

export const usePredict = routeLoader$(async () => {
    const ws = new WebSocket('ws://localhost:8080/predict?url=' + data.url);

    ws.on('open', () => {
        console.log('Connected to server');
    });

    ws.on('message', (message: string) => {
        console.log(`Received message from server: ${message}`);
        list.push({ text: message })
    });

    ws.on('close', () => {
        console.log('Disconnected from server');
    });
    console.log("data -> ", data);
});

export const head: DocumentHead = {
    title: "Predict",
    meta: [
        {
            name: "description",
            content: "Looking for tacos",
        },
    ],
};
