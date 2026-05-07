# Computacao em Nuvem: Uma Analise Comparativa das Plataformas (AWS, GCP, Azure)

**Autor:** ARAUJO, Mikael Nilton de
**Ano:** 2019
**Fonte:** TCC (Tecnologia em ADS), IFRN Campus Pau dos Ferros
**Orientador:** Me. Jeferson Queiroga Pereira
**PDF:** `ifrn_cloud_comparison_tcc.pdf`

## Resumo

Trabalho comparativo das tres principais plataformas de computacao em nuvem: Amazon Web Services (AWS), Google Cloud Platform (GCP) e Microsoft Azure. Aborda a arquitetura de hardware/software de cada plataforma, principais servicos oferecidos (computacao, armazenamento, rede, banco de dados), modelos de servico (IaaS, PaaS, SaaS), modelos de implantacao (publica, privada, hibrida, comunitaria), e aspectos como preco, presenca de mercado, seguranca e tratamento de dados. Conclui com estudo comparativo de pontos comuns entre as tres ofertas.

## Analise

Trabalho de fundamentacao teorica abrangente sobre computacao em nuvem no contexto brasileiro. Embora nao aborde diretamente metadata services ou cloud-init, fornece o contexto de como os tres grandes provedores estruturam seus servicos — cada um deles possui seu proprio IMDS (AWS IMDS, GCP Metadata Server, Azure IMDS). O trabalho pode ser referenciado para fundamentacao sobre modelos de servico em nuvem e para contextualizar por que provedores menores/plataformas privadas como Incus precisam implementar servicos equivalentes.

## Pontos-chave

- Computacao em nuvem: conceito desde 1960 (time-sharing de McCarthy), termo cunhado em 1997 por Chellappa
- Classificacao: nuvem publica, privada, hibrida e comunitaria
- Modelos de servico: IaaS (hardware virtualizado), PaaS (plataforma de desenvolvimento), SaaS (software pronto)
- AWS: lider de mercado, lancada em 2006, 24 regioes, 84 AZs, SLA de <6min downtime/ano
- GCP: infraestrutura do Google, forte em ML/BigData, menor presenca global que AWS
- Azure: integracao com ecossistema Microsoft, forte em mercado enterprise
- Todos oferecem: VPC, computacao elastica, armazenamento em blocos/objetos, bancos gerenciados

## Informacoes Pertinentes para o TCC

- **Contexto dos provedores:** cada provedor (AWS, GCP, Azure) tem seu proprio metadata service (IMDS) — seu TCC cria o equivalente para Incus
- **Definicoes de nuvem privada:** o trabalho define nuvem privada como "infraestrutura ambientada dentro da propria empresa, com ambiente mais controlado e seguro" — exatamente o cenario do seu TCC
- **Modelos IaaS/PaaS:** seu metadata service opera na camada IaaS, fornecendo dados de inicializacao que sao padrao em provedores publicos mas ausentes no Incus
- **Referencia brasileira:** pode citar como fundamentacao para conceitos basicos de computacao em nuvem escritos em portugues

## Sugestao BibTeX

```bibtex
@monography{araujo2019computacao,
  title={Computa{\c{c}}{\~a}o em Nuvem: Uma An{\'a}lise Comparativa das Plataformas Disponibilizadas por Amazon, Google e Microsoft},
  author={Ara{\'u}jo, Mikael Nilton de},
  year={2019},
  school={Instituto Federal de Educa{\c{c}}{\~a}o, Ci{\^e}ncia e Tecnologia do Rio Grande do Norte},
  address={Pau dos Ferros},
  type={Trabalho de Conclus{\~a}o de Curso (Tecnologia em ADS)}
}
```
