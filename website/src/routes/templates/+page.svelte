<script>
	export let data;
	const templates = data;
</script>

<div class="flex justify-center w-full">
	<div class="flex flex-col max-w-lg lg:max-w-3xl">
		<div class="select-none text-xl lg:text-3xl w-full text-center font-bold mt-4">模板提名</div>
		<div class="flex flex-row gap-2 mt-1">
			<input
				type="text"
				placeholder="输入含模板的词条"
				class="grow input input-bordered input-primary"
			/>
			<button class="btn btn-primary">测试词条</button>
		</div>
		<div class="select-none text-xl lg:text-3xl w-full text-center font-bold mt-4">模板列表</div>
		<div class="flex flex-row justify-center flex-wrap gap-2 max-w-max">
			{#each Object.entries(templates) as [name, content]}
				<div
					class="flex-none has-tooltip min-w-min border rounded-full border-primary overflow-hidden"
				>
					<div class="mt-12 tooltip min-w-min max-w-lg rounded-xl p-1 bg-primary flex flex-row justify-center flex-wrap gap-1">
						{#each content.values as v}
							<div class="border border-neutral rounded-md text-neutral text-sm pl-1 pr-1">{v}</div>
						{/each}
					</div>
					<div class="bg-primary text-neutral w-full text-md font-bold pl-2 pr-2">{'{{'}{name}{'}}'}</div>
					<div class="text-sm text-center pl-2 pr-2">{content.desc}</div>
					<!-- <div class="collapse-content">
						{#each content.values as v}
							<div class="badge badge-secondary badge-outline m-1">{v}</div>
						{/each}
					</div> -->
				</div>
			{/each}
		</div>
		<!-- <div class="grid lg:grid-cols-2 gap-4 max-w-lg lg:max-w-3xl">
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
		</div> -->
	</div>
</div>
