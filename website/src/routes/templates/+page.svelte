<script>
	export let data;
	const templates = data;
</script>

<div class="flex justify-center w-full h-full">
	<div class="grid content-center max-w-lg lg:max-w-3xl ml-4 mr-4 lg:m-0">
		<div class="">
			<div class="select-none text-3xl w-full text-center font-bold">模板测试</div>
			<div class="flex flex-row gap-2 mt-2 lg:pl-4 lg:pr-4">
				<input
					type="text"
					placeholder="输入含模板的词条"
					class="grow input input-bordered input-primary"
				/>
				<button class="btn btn-primary">试一试</button>
			</div>
		</div>
		<div class="divider mt-4 mb-4"></div>
		<div class="">
			<div class="select-none text-3xl w-full text-center font-bold">模板列表</div>
			<div class="flex flex-row justify-center flex-wrap gap-2 max-w-max mt-2">
				{#each Object.entries(templates) as [name, content]}
					<div
						class="flex-none has-tooltip min-w-min border rounded-full border-primary overflow-hidden"
					>
						<div
							class="mt-12 tooltip min-w-min max-w-lg rounded-xl p-1 bg-primary flex flex-row justify-center flex-wrap gap-1"
						>
							{#each content.values as v}
								<div class="border border-neutral rounded-md text-neutral text-sm pl-1 pr-1">
									{v}
								</div>
							{/each}
						</div>
						<div class="bg-primary text-neutral w-full text-md font-bold pl-2 pr-2">
							{'{{'}{name}{'}}'}
						</div>
						<div class="text-sm text-center pl-2 pr-2">{content.desc}</div>
					</div>
				{/each}
			</div>
		</div>
		<div class="mt-20"></div>
	</div>
</div>
