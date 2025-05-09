# Multi-GPU Records

## Add a PUF component

**Overview.** The PUF component in `puf.go` has been added into the simulated R9 Nano GPU. It is connected into the command processor with the port `ToCP` and `ToPUF`. In the test file `puf_test/main.go`, the GPU driver can fullfill a simple challenge-response communication with the created PUF instance, with the command processor being the intermediary.


**Implementation details.**
- `samples/runner/puf/puf.go`: the main file creating a hash-simulated PUF component. It is developed based on code of DMA, PMC and other component. It is a `TickingComponent` with several input parameters.
- `samples/runner/puf_test`: the test file where we invoke the modified GPU initilization function of the framework. Currently, we let the GPU driver to send a challenge request to the command processor (CP). The CP then forwards the request to the connected PUF instance. The connection between CP and PUF is similar to the existing component like DMA, PMC, etc. Note that PUF only have a `ToCP` local port and we add an extra port `ToPUF` for CP. There is no remote external port like RDMA for PUF. Finally, PUF handles the challenge and sends back the response all the way to the driver.
- modified files for PUF-related functions:
    - `samples/runner/r9nanobuilder.go`
    - `samples/runner/platform.go`
    - `timing/cp/builder.go`
    - `timing/cp/commandprocessor.go`

**Integrate PUF into multi-GPU model**. The same manners can be direcly applied in the multi-GPU model. That is, one GPU driver sends the challenge request to the PUF instance one-by-one through each command processor. (The parallezation of these PUFs is not investigated yet.)


## TODO
- [Done] ~~Creating multiple PUF instances for multiple GPU instances.~~
    - Initialize the PUF component with a unique nonce-like input
    - Generate the response differently. Maybe using the hash function with an input nonce can simulate the different mappings for different PUF instances.
    - Make sure the each PUF instance is well plugged inside its corresponding GPU.

- [Done] ~~Test for multiple PUFs on different GPUs.~~


- The parallezation of multiple PUF instances.
    - The GPU driver sends the challenge requests to the command processors on each GPU. The challenges can be the same or different. 
    - The PUF instance generate its response according to its own challenge. Then it sends back the response to the command processor.
    - The generating (calling) opertations of each PUF happens at most the same time, i.e., the parallezation. So the driver should receive multiple responses from each PUF (or command processor) at most the same time.


## Remote Attestation



## Memory model

### Prompt record

> Please analyze this codebase (akita and MGPUSim) in details. Summarize the multi-GPU-related part, especially focusing on its memory model and interconnect model.   


> I used to think that RDMA-based direct communication has nothing to do with the page migration. Now based on your analysis, it seems the migration must be done with RDMA. Is that correct or not? And what migration techniques does MGPUSim use? I notice that there are plenty of them like first-touch migration, on-touch migration, access counter-based migration and so on. Which technique does MGPUSim use and how does it implement it?

> I learned that except for page migration, direct cache access is another technique that enables GPU-to-GPU communication with smaller data granularity. In many acedamic papers, this technique is more often combined with the RDMA instead of the page migration. However, MGPUSim seems not to implement direct cache access. So there is only one direct GPU-to-GPU memory communication technique here, i.e., the RDMA-based page migration. Am I correct? Please deeply analyze this codebase. And you can combine @Web  to give a mroe reliable answer.

> So in summary, can you show me a complete workflow about how MGPUSim handles a multi-GPU workload using the page migration technique and the unified memory model. Please explain in details. Focus on how the migration and unified model is used.

> Analyze this MGPUSim simulator in detail. Focus on its overall structure, workflow and involved component. I wish to add the remote attestation mechanism in the current codebase. Specifically, the basic process is: the GPU driver will act as an intermediate to forward the attestation request from the user-side script. Then it request the GPU who returns a signed measurement and form an attestation report to the driver. The driver or the user-side script can verify the authenticity of the report and measurement. Then compares the measurement with a reference "golden" value. You can combine the knowledge of remote attestation on @Web  and take a Nvidia H100 Attestation scheme @https://github.com/NVIDIA/nvtrust  as an example. You only need to provide a minimum skeleton of the attestation procedure in the MGPUSim framework. No real bussiness logic is required for now.
