<script lang="ts">
  import { tr } from '../../application/i18n';
</script>

<section class="features">
  <article class="card feature">
    <span class="feature-icon magenta">⌕</span>
    <h3>{$tr('welcome.feature1Title')}</h3>
    <p class="muted">{$tr('welcome.feature1Body')}</p>
  </article>
  <article class="card feature">
    <span class="feature-icon cyan">↑</span>
    <h3>{$tr('welcome.feature2Title')}</h3>
    <p class="muted">{$tr('welcome.feature2Body')}</p>
  </article>
  <article class="card feature">
    <span class="feature-icon lime">◎</span>
    <h3>{$tr('welcome.feature3Title')}</h3>
    <p class="muted">{$tr('welcome.feature3Body')}</p>
  </article>
</section>

<section class="how" id="how">
  <h2>{$tr('welcome.howTitle')}</h2>
  <ol>
    <li><span>1</span>{$tr('welcome.howStep1')}</li>
    <li><span>2</span>{$tr('welcome.howStep2')}</li>
    <li><span>3</span>{$tr('welcome.howStep3')}</li>
  </ol>
</section>

<section class="privacy card" id="privacy">
  <div>
    <h2>{$tr('welcome.privacyTitle')}</h2>
    <p class="muted">{$tr('welcome.privacyBody')}</p>
  </div>
  <p class="tiny dim">{$tr('welcome.youtubeNote')}</p>
</section>

<style>
  .features {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
    padding: 1rem 0 3.5rem;
  }

  .feature {
    padding: 1.3rem;
  }

  .feature-icon {
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    border-radius: 13px;
    margin-bottom: 1rem;
    font-size: 1.3rem;
    font-weight: 800;
  }

  .feature-icon.magenta {
    background: rgb(255 45 149 / 0.16);
    color: var(--magenta);
  }

  .feature-icon.cyan {
    background: rgb(34 211 238 / 0.12);
    color: var(--cyan);
  }

  .feature-icon.lime {
    background: rgb(182 255 59 / 0.1);
    color: var(--lime);
  }

  .feature h3 {
    margin-bottom: 0.4rem;
    font-size: 1.05rem;
  }

  .feature p {
    font-size: 0.9rem;
    line-height: 1.55;
  }

  .how {
    padding: 2rem 0 3rem;
  }

  .how h2,
  .privacy h2 {
    margin-bottom: 1rem;
    font-size: 1.4rem;
  }

  .how ol {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .how li {
    display: flex;
    gap: 0.8rem;
    align-items: flex-start;
    color: var(--muted);
    line-height: 1.5;
    font-size: 0.92rem;
  }

  .how li span {
    flex: none;
    display: grid;
    place-items: center;
    width: 30px;
    height: 30px;
    border-radius: 50%;
    border: 1px solid var(--border-bright);
    color: var(--cyan);
    font-family: var(--mono);
    font-size: 0.8rem;
  }

  .privacy {
    padding: 1.6rem;
    display: grid;
    gap: 1rem;
    margin-bottom: 3rem;
  }

  .privacy p {
    line-height: 1.6;
    max-width: 80ch;
    margin-top: 0.5rem;
  }

  @media (max-width: 900px) {
    .features,
    .how ol {
      grid-template-columns: 1fr;
    }
  }
</style>
