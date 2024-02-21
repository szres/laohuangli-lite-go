<script>
	export let data;
	const templates = data;
</script>

<h2>模板词条</h2>

{#each Object.entries(templates) as [name, content]}
	<p>{'{{'}{name}{'}}'}:{content.desc}</p>
	{#each content.values as v}
		<span>{v}</span>
	{/each}
{/each}
