<script>
	export let data;
	const caches = data;
</script>

<div class="flex flex-col">
	<div class="select-none text-lg w-full text-center font-bold mt-4">今日指引</div>
	<div class="stats stats-horizontal shadow">
		<div class="stat text-center">
			<div class="stat-title">穿搭</div>
			<div class="stat-value">-</div>
			<div class="stat-desc">Coming Soon</div>
		</div>

		<div class="stat text-center">
			<div class="stat-title">饮食</div>
			<div class="stat-value">-</div>
			<div class="stat-desc">Coming Soon</div>
		</div>

		<div class="stat text-center">
			<div class="stat-title">出行</div>
			<div class="stat-value">-</div>
			<div class="stat-desc">Coming Soon</div>
		</div>
	</div>
</div>
<div class="select-none self-start text-xl w-full text-center font-bold mt-4">我的老黄历</div>
<div class="text-center">Coming Soon...</div>

<div class="select-none self-start text-xl w-full text-center font-bold mt-4">众生的老黄历</div>
<div class="overflow-x-auto">
	<table class="table table-zebra">
		<thead>
			<tr>
				<th>{caches.date}</th>
				<th>老黄历</th>
			</tr>
		</thead>
		<tbody>
			{#each Object.entries(caches.caches) as [name, content]}
				<tr>
					<td>{content.name}</td>
					<td>{content.result}</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
