<script>
	export let data;
	const caches = data;
</script>

<div class="flex justify-center w-full">
	<div class="flex flex-col max-w-lg lg:max-w-3xl">
		<div class="select-none text-xl lg:text-3xl w-full text-center font-bold mt-4">今日指引</div>
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

		<div class="select-none text-xl lg:text-3xl w-full text-center font-bold mt-4">众生的老黄历</div>
		<div class="overflow-x-auto">
			<table class="table table-sm table-zebra">
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
	</div>
</div>
