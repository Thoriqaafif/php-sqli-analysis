<template>
    <div class="flex flex-col w-full items-center">
        <h1 class="text-lg font-sans font-medium">Source</h1>
        <div class="w-full bg-green-300 text-center rounded-sm">
            <h1 class="font-semibold">{{ trace.taint_source.location.path }}</h1>
            <p>Line: {{ trace.taint_source.location.start.line }}</p>
            <p class="text-lg font-semibold">"{{ trace.taint_source.content }}"</p>
        </div>
        <v-icon name="fa-long-arrow-alt-down" class="w-[2.5rem] h-[2.5rem] text-blue-400"></v-icon>

        <h1 class="text-lg font-sans font-medium">Intermediate</h1>
        <div v-for="(interm, index) in trace.intermediate_vars" :key="index" class="w-full text-center">
            <div class="w-full bg-yellow-300 text-center rounded-sm">
                <h1 class="font-semibold">{{ interm.location.path }}</h1>
                <p>Line: {{ interm.location.start.line }}</p>
                <p class="text-lg font-semibold">"{{ interm.content }}"</p>
            </div>
            <v-icon name="fa-long-arrow-alt-down" class="w-[2.5rem] h-[2.5rem] text-blue-400"></v-icon>
        </div>

        <h1 class="text-lg font-sans font-medium">Sink</h1>
        <div class="w-full bg-red-300 text-center rounded-sm">
            <h1 class="font-semibold">{{ trace.taint_sink.location.path }}</h1>
            <p>Line: {{ trace.taint_sink.location.start.line }}</p>
            <p class="text-lg font-semibold">"{{ trace.taint_sink.content }}"</p>
        </div>
    </div>
</template>

<script>
export default {
    props: {
        trace: {},
    }
}
</script>