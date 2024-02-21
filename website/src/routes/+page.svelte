
<script>
	export let data;
	const caches = data;
</script>


<h2>今日指引</h2>
<h2>我的老黄历</h2>

<h2>别人的老黄历</h2>

<p>{caches.date}</p>
{#each Object.entries(caches.caches) as [name, content]}
<p>{content.name}:{content.result}</p>
{/each}
