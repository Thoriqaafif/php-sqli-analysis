<template>
    <div class="h-100vh">
        <header class="h-[25vh]">
            <div
                class="flex justify-between h-[40%] px-5 py-4 border border-x-0 border-t-0 border-b-1 border-slate-400">
                <h1 class="text-xl font-mono font-bold">Your Projects</h1>
            </div>
            <div class="flex h-[60%] justify-between px-5 py-3">
                <div class="relative my-auto w-[40%]">
                    <input v-model="search"
                        class="w-full bg-transparent placeholder:text-slate-400 text-slate-700 text-sm border border-slate-200 rounded-lg pl-3 pr-28 py-2 transition duration-300 ease focus:outline-none focus:border-slate-400 hover:border-slate-300 shadow-sm focus:shadow"
                        placeholder="Search ..." />
                </div>
                <div class="my-auto mx-5">
                    <button class="bg-red-400 rounded-lg py-1 px-5" @click="deleteProjects">
                        <p class="font-mono text-md font-semibold">Delete</p>
                    </button>
                </div>
            </div>
        </header>
        <main class="h-[75vh] p-0 overflow-y-scroll">
            <table class="min-w-full table-auto p-0 m-0">
                <thead class="bg-slate-300 sticky top-0">
                    <tr>
                        <th class="px-4"></th>
                        <th class="w-1/4 px-4  text-left font-mono">Name</th>
                        <th class="px-4 text-left font-mono">PHP File Counts</th>
                        <th class="px-4 text-left font-mono">Detected Vulnerabilities</th>
                        <th class="px-4 text-left font-mono">Scan Time</th>
                        <th class="px-4 text-left font-mono">Scanned At</th>
                        <th class="px-4 text-left font-mono">Result</th>
                    </tr>
                </thead>
                <div v-if="error">{{ error }}</div>
                <tbody v-else>
                    <tr v-for="(project, index) in searchedProjects" :key="project.id" :class="rowClass(index)">
                        <td class="px-4 py-2 top-0">
                            <input type="checkbox" @click="selectProject(index)" :checked="project.selected">
                        </td>
                        <td class="px-4 py-2 text-left font-mono">
                            <router-link :to="'/detail/' + project.id">
                                {{ project.name }}
                            </router-link>
                        </td>
                        <td class="px-4 py-2 font-mono">{{ project.num_of_files }}</td>
                        <td class="px-4 py-2 font-mono">{{ project.detected_vulns }}</td>
                        <td class="px-4 py-2 font-mono">{{ project.scan_time }}</td>
                        <td class="px-4 py-2 font-mono">{{ (new Date(project.created_at)).toUTCString() }}</td>
                        <td class="px-4 py-2 font-mono"><router-link :to="'/detail/' + project.id">
                                <v-icon name="co-arrow-thick-from-left"></v-icon>
                            </router-link></td>
                    </tr>
                </tbody>
            </table>
        </main>
    </div>
</template>

<script>
export default {
    computed: {
        searchedProjects() {
            return this.projects
                .filter((p) => p.name.toLowerCase().startsWith(this.search.toLowerCase()))
        }
    },
    data() {
        return {
            projects: [],
            error: null,
            search: "",
        }
    },
    methods: {
        rowClass(idx) {
            if (idx % 2 == 1) {
                return "bg-gray-100"
            }
            return "bg-white"
        },
        selectProject(idx) {
            this.projects[idx].selected = !this.projects[idx].selected
        },
        async deleteProjects() {
            for (const p of this.projects) {
                if (p.selected) {
                    await this.deleteProject(p.id)
                }
            }
            this.fetchProjects()
        },
        async deleteProject(id) {
            const apiUrl = `http://localhost:8080/api/project/delete/${id}`;
            try {
                const response = await fetch(apiUrl, {
                    method: 'DELETE',
                });

                if (response.ok) {
                    console.log('Item deleted successfully');
                } else {
                    console.error('Failed to delete item', response.status);
                }
            } catch (error) {
                console.error('Error:', error);
            }
        },
        async fetchProjects() {
            // fetch projects data
            try {
                const response = await fetch('http://localhost:8080/api/project');
                if (!response.ok) throw new Error('Failed to fetch data');
                this.projects = await response.json();
                // add 'selected' property
                this.projects.forEach((p) => p.selected = false)
            } catch (err) {
                this.error = err
            }
        }
    },
    beforeMount() {
        this.fetchProjects()
    }
}
</script>