# Computacao em Nuvem no Contexto das Smart Grids

**Autor:** SOUSA, Jeovane Vicente de
**Ano:** 2018
**Fonte:** Tese de Doutorado, Escola de Engenharia de Sao Carlos, USP
**Orientador:** Prof. Titular Denis Vinicius Coury
**PDF:** `usp_cloud_lxd_2018.pdf`

## Resumo

Tese de doutorado (183 paginas) que desenvolve uma infraestrutura de computacao em nuvem privada para smart grids utilizando ferramentas open source. Utiliza OpenNebula como gerenciador de infraestrutura com hipervisor KVM para VMs e LXD para containers Linux. A infraestrutura e single-server com alocacoes estaticas (VMs KVM) e dinamicas (containers LXD para inicializacao rapida sob demanda). Implementa servicos de banco de dados (sMAP para time-series, MongoDB), um Head-End Server, e uma aplicacao de localizacao de faltas usando mineracao de dados (DAMICORE). A aplicacao foi validada com centenas de simulacoes de falta, reduzindo a multipla estimacao em mais de 80%.

## Analise

Esta tese e relevante por ser um dos poucos trabalhos academicos brasileiros que utiliza LXD em um contexto de nuvem privada. O autor utiliza LXD especificamente pela vantagem de inicializacao rapida de containers comparada a VMs, mas nao aborda a questao de metadata services ou cloud-init para configuracao automatica — justamente a lacuna que o seu TCC preenche. A tese mostra que a plataforma LXD/Incus e viavel para uso academico e de pesquisa, validando o contexto em que seu servico de metadados opera.

## Pontos-chave

- Utiliza LXD + OpenNebula + KVM em configuracao single-server para nuvem privada
- LXD escolhido por inicializacao rapida (~1s) vs VMs KVM (~30s)
- Integracao via add-on LXDoNe para OpenNebula gerenciar containers LXD
- Containers LXD instanciados sob demanda para servicos que nao precisam estar sempre ativos
- Infraestrutura implementada com ferramentas 100% open source
- Compara Docker (aplicacoes), LXD (SO completo), Kubernetes (orquestracao)

## Informacoes Pertinentes para o TCC

- **Validacao da plataforma LXD em contexto academico:** demonstra que LXD e usado em pesquisa brasileira, mas sem metadata service — reforca a lacuna que seu TCC preenche
- **Inicializacao de containers:** a tese menciona scripts de inicializacao definidos nos templates de VM do OpenNebula (Figura 24-25), que e exatamente o tipo de configuracao que cloud-init + metadata service automatizaria
- **Citar para:** contextualizar o uso de LXD/containers em nuvem privada no Brasil, mostrar que a plataforma e viavel mas carece de ferramentas de inicializacao automatica padronizadas
- **Contraste:** OpenNebula tem integracao com cloud-init via contextualizacao; Incus (fork do LXD) nao tem metadata service HTTP equivalente

## Sugestao BibTeX

```bibtex
@phdthesis{sousa2018computacao,
  title={Computa{\c{c}}{\~a}o em Nuvem no Contexto das Smart Grids: Uma Aplica{\c{c}}{\~a}o para Aux{\'\i}lio {\`a} Localiza{\c{c}}{\~a}o de Faltas em Sistemas de Distribui{\c{c}}{\~a}o},
  author={Sousa, Jeovane Vicente de},
  year={2018},
  school={Escola de Engenharia de S{\~a}o Carlos, Universidade de S{\~a}o Paulo},
  address={S{\~a}o Carlos},
  type={Tese de Doutorado}
}
```
