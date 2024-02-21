<script>
	export let data;
	const templates = data;
</script>

{#each Object.entries(templates) as [name, content]}
<div class="collapse collapse-arrow border border-base-300 bg-base-200 mt-1">
	<input type="checkbox" /> 
	<div class="collapse-title text-lg font-bold">
		<div class="badge badge-accent">{'{{'}{name}{'}}'}</div>
		<div class="text-sm">{content.desc}</div>
	</div>
	<div class="collapse-content">
		{#each content.values as v}
		<div class="badge badge-secondary badge-outline m-1">{v}</div>
		{/each}
	</div>
</div>
{/each}
